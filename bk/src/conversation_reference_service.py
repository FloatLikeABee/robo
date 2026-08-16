"""
Rolling conversation-reference summaries for multi-turn modules (dialogue, agents, etc.).

When enabled in system settings, after each meaningful exchange a background thread asks the
**system default** LLM (not the module's own model) to append key-point bullets to a cached
summary scoped per conversation session. The cache is injected into subsequent prompts as light
context — independent of RAG/tools/model used for the main reply.
"""
from __future__ import annotations

import logging
import os
import shutil
import threading
from datetime import datetime, timezone
from typing import Any, Dict, Optional, TYPE_CHECKING

from tinydb import TinyDB, Query

from .config import settings as env_settings

if TYPE_CHECKING:
    from .system_settings_manager import SystemSettingsManager

logger = logging.getLogger(__name__)

_ssm_singleton: Optional["SystemSettingsManager"] = None


def get_ssm() -> "SystemSettingsManager":
    """Shared SystemSettingsManager for modules that are not wired through RAGAPI."""
    global _ssm_singleton
    if _ssm_singleton is None:
        from .system_settings_manager import SystemSettingsManager

        _ssm_singleton = SystemSettingsManager()
    return _ssm_singleton

MAX_SUMMARY_CHARS = 8000
EXCERPT_MAX = 3500


def _truncate(text: str, max_len: int) -> str:
    text = (text or "").strip()
    if len(text) <= max_len:
        return text
    return text[: max_len - 3] + "..."


def dialogue_scope(dialogue_id: str, conversation_id: str) -> str:
    return f"dialogue:{dialogue_id}:{conversation_id}"


def agent_scope(agent_id: str, session_id: Optional[str]) -> str:
    return f"agent:{agent_id}:{session_id or 'default'}"


def assistant_scope(assistant_id: str, session_id: Optional[str]) -> str:
    return f"assistant:{assistant_id}:{session_id or 'default'}"


def adviser_scope(adviser_id: str, session_id: Optional[str]) -> str:
    return f"assistant:{adviser_id}:{session_id or 'default'}"


def customization_scope(profile_id: str, session_id: Optional[str]) -> str:
    return f"assistant:{profile_id}:{session_id or 'default'}"


class ConversationReferenceStore:
    """TinyDB-backed store for per-scope summary text.

    Uses a lock because background threads call ``set_summary`` while request handlers
    call ``get_summary``; TinyDB's default JSON file is not safe for concurrent writers
    and can end up empty or truncated (JSONDecodeError on read).
    """

    def __init__(self) -> None:
        os.makedirs(env_settings.data_directory, exist_ok=True)
        self._path = os.path.join(env_settings.data_directory, "conversation_references.json")
        self._lock = threading.Lock()
        self._open_db()

    def _open_db(self) -> None:
        self.db = TinyDB(self._path)
        self.query = Query()

    def _recover_corrupt_db(self) -> None:
        """Replace a missing/invalid DB file so reads never stay broken."""
        try:
            self.db.close()
        except Exception:
            pass
        stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
        backup = f"{self._path}.corrupt.{stamp}"
        try:
            if os.path.isfile(self._path):
                shutil.move(self._path, backup)
                logger.warning(
                    "Moved corrupt conversation_references store to %s; starting fresh",
                    backup,
                )
        except OSError as e:
            logger.warning("Could not move corrupt DB (%s); trying remove", e)
            try:
                os.remove(self._path)
            except OSError:
                pass
        self._open_db()

    def get_summary(self, scope: str) -> str:
        with self._lock:
            for attempt in range(2):
                try:
                    doc = self.db.get(self.query.scope == scope)
                    if not doc:
                        return ""
                    return str(doc.get("summary_text") or "")
                except Exception as e:
                    logger.warning(
                        "conversation_references read failed (attempt %s): %s",
                        attempt + 1,
                        e,
                    )
                    if attempt == 0:
                        self._recover_corrupt_db()
                    else:
                        return ""
            return ""

    def set_summary(self, scope: str, text: str) -> None:
        text = (text or "").strip()
        with self._lock:
            for attempt in range(2):
                try:
                    if not text:
                        self.db.remove(self.query.scope == scope)
                        return
                    self.db.upsert(
                        {
                            "scope": scope,
                            "summary_text": text[:MAX_SUMMARY_CHARS],
                            "updated_at": datetime.now(timezone.utc).isoformat(),
                        },
                        self.query.scope == scope,
                    )
                    return
                except Exception as e:
                    logger.warning(
                        "conversation_references write failed (attempt %s): %s",
                        attempt + 1,
                        e,
                    )
                    if attempt == 0:
                        self._recover_corrupt_db()
                    else:
                        return


_store = ConversationReferenceStore()


def _merge_summaries(prior: str, new_block: str) -> str:
    prior = (prior or "").strip()
    new_block = (new_block or "").strip()
    combined = (prior + "\n\n" + new_block).strip() if prior else new_block
    if len(combined) <= MAX_SUMMARY_CHARS:
        return combined
    # Keep the tail: drop lines from the top until under cap
    lines = combined.splitlines()
    out: list[str] = []
    total = 0
    for line in reversed(lines):
        if total + len(line) + 1 > MAX_SUMMARY_CHARS:
            continue
        out.append(line)
        total += len(line) + 1
    return "\n".join(reversed(out)).strip()


def _create_caller_for_system_defaults(ssm: "SystemSettingsManager") -> Any:
    """LLM caller using persisted system default provider/model (env API keys)."""
    from .llm_factory import LLMFactory, LLMProvider

    s = ssm._load_settings()
    provider_str = (s.default_llm_provider or env_settings.default_llm_provider or "gemini").lower().strip()
    model_name = s.default_model or env_settings.default_model

    if provider_str == "gemini":
        api_key = env_settings.gemini_api_key
        model = model_name or env_settings.gemini_default_model
        provider = LLMProvider.GEMINI
    elif provider_str == "qwen":
        api_key = env_settings.qwen_api_key
        model = model_name or env_settings.qwen_default_model
        provider = LLMProvider.QWEN
    elif provider_str == "mistral":
        api_key = env_settings.mistral_api_key
        model = model_name or env_settings.mistral_default_model
        provider = LLMProvider.MISTRAL
    elif provider_str == "groq":
        api_key = getattr(env_settings, "groq_api_key", "") or ""
        model = model_name or getattr(env_settings, "groq_default_model", "llama-3.3-70b-versatile")
        provider = LLMProvider.GROQ
    else:
        api_key = env_settings.gemini_api_key
        model = model_name or env_settings.gemini_default_model
        provider = LLMProvider.GEMINI

    if not api_key:
        raise ValueError(f"No API key configured for system default provider '{provider_str}'")

    return LLMFactory.create_caller(
        provider=provider,
        api_key=api_key,
        model=model,
        temperature=0.2,
        max_tokens=2048,
    )


def _summarize_exchange(
    ssm: "SystemSettingsManager",
    prior_summary: str,
    user_message: str,
    assistant_message: str,
) -> str:
    caller = _create_caller_for_system_defaults(ssm)
    prompt = (
        "You consolidate conversation memory as short bullet points for future context.\n"
        "Given PRIOR NOTES (may be empty) and the latest USER / ASSISTANT exchange, "
        "output ONLY new or updated bullet lines (max 8 lines) capturing decisions, facts, "
        "constraints, and open questions. No preamble. Use '-' bullets.\n\n"
        f"PRIOR NOTES:\n{_truncate(prior_summary, EXCERPT_MAX)}\n\n"
        f"USER:\n{_truncate(user_message, EXCERPT_MAX)}\n\n"
        f"ASSISTANT:\n{_truncate(assistant_message, EXCERPT_MAX)}\n"
    )
    text = caller.generate(prompt)
    return (text or "").strip()


def get_reference_system_message(ssm: "SystemSettingsManager", scope: str) -> str:
    """Returns text for a system message, or empty string if disabled or empty."""
    s = ssm._load_settings()
    if not getattr(s, "conversation_reference_cache_enabled", False):
        return ""
    summary = _store.get_summary(scope)
    if not summary.strip():
        return ""
    return (
        "### Conversation reference (cached key points)\n"
        f"{summary.strip()}\n"
        "Use only if helpful; it is a compact memory, not full history."
    )


def get_reference_prefix_block(ssm: "SystemSettingsManager", scope: str) -> str:
    """Plaintext block to prepend to a single-shot prompt (agents/customizations)."""
    msg = get_reference_system_message(ssm, scope)
    return f"{msg}\n\n" if msg else ""


def schedule_conversation_reference_update(
    ssm: "SystemSettingsManager",
    scope: str,
    user_message: str,
    assistant_message: str,
) -> None:
    """Fire-and-forget background thread: merge summarized exchange into store."""

    def worker() -> None:
        try:
            s = ssm._load_settings()
            if not getattr(s, "conversation_reference_cache_enabled", False):
                return
            min_c = getattr(s, "conversation_reference_min_exchange_chars", 200)
            if len(user_message or "") + len(assistant_message or "") < min_c:
                return
            prior = _store.get_summary(scope)
            new_part = _summarize_exchange(ssm, prior, user_message, assistant_message)
            if not new_part:
                return
            merged = _merge_summaries(prior, new_part)
            _store.set_summary(scope, merged)
        except Exception:
            logger.exception("conversation_reference update failed for scope=%s", scope)

    t = threading.Thread(target=worker, daemon=True, name="conv-ref-update")
    t.start()
