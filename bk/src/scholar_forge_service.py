"""
ScholarForge orchestration — dual-model MCP-style academic composition pipeline.

Organizer model: clarifies requirements, plans structure, summarizes sections.
Writer model: produces section content paragraph-by-paragraph.
Reviewer model: critiques each paragraph and returns structured feedback.
Writer revises based on reviewer report until approved or max rounds.
Image model: describes uploaded figures for placement.
"""
from __future__ import annotations

import base64
import io
import json
import logging
import os
import re
import threading
import uuid
from dataclasses import dataclass
from typing import Any, Dict, Iterator, List, Optional, Tuple
from urllib.parse import urlparse

import requests

from .config import settings
from .llm_factory import LLMFactory, LLMProvider
from .llm_langchain_wrapper import LangChainLLMWrapper
from .models import (
    LLMProviderType,
    ScholarForgeClarificationForm,
    ScholarForgeFormField,
    ScholarForgeMaterialInput,
    ScholarForgeParagraphRecord,
    ScholarForgePipelineStep,
    ScholarForgeProfile,
    ScholarForgeReviewReport,
    ScholarForgeSection,
    ScholarForgeStatus,
    ScholarForgeStructure,
)
from .scholar_forge_pdf import markdown_to_pdf, save_markdown
from .scholar_forge_pdf import markdown_to_pdf, save_markdown
from .scholar_forge_rag import ingest_materials_to_rag, query_project_rag
from .scholar_forge_enhanced import (
    ROLE_POLISHER,
    build_polisher_prompt,
    enhanced_clarification_prompt,
    enhanced_reviewer_prompt,
    enhanced_revise_prompt,
    enhanced_structure_prompt,
    enhanced_writer_paragraph_prompt,
    extract_polished_document,
    gather_references,
)

logger = logging.getLogger(__name__)

ROLE_ORGANIZER = "organizer"
ROLE_WRITER = "writer"
ROLE_REVIEWER = "reviewer"
ROLE_IMAGE = "image"

PREV_MAX_CHARS = 20_000
RAG_QUERY_RESULTS = 5
KEEPALIVE_INTERVAL_SEC = 8

_PROVIDER_ORDER = ("gemini", "qwen", "groq", "mistral")


@dataclass(frozen=True)
class _RoleLLMConfig:
    api_key: str
    base_url: str
    model: str
    provider: str


def _role_llm_config(role: str) -> _RoleLLMConfig:
    if role == ROLE_ORGANIZER:
        return _RoleLLMConfig(
            api_key=(settings.scholar_forge_organizer_api_key or "").strip(),
            base_url=(settings.scholar_forge_organizer_base_url or "").strip().rstrip("/"),
            model=(settings.scholar_forge_organizer_model or "").strip(),
            provider=(settings.scholar_forge_organizer_provider or settings.default_llm_provider or "gemini").lower(),
        )
    if role == ROLE_WRITER:
        return _RoleLLMConfig(
            api_key=(settings.scholar_forge_writer_api_key or "").strip(),
            base_url=(settings.scholar_forge_writer_base_url or "").strip().rstrip("/"),
            model=(settings.scholar_forge_writer_model or "").strip(),
            provider=(settings.scholar_forge_writer_provider or settings.default_llm_provider or "gemini").lower(),
        )
    if role == ROLE_REVIEWER:
        reviewer_key = (settings.scholar_forge_reviewer_api_key or settings.scholar_forge_organizer_api_key or "").strip()
        reviewer_url = (settings.scholar_forge_reviewer_base_url or settings.scholar_forge_organizer_base_url or "").strip().rstrip("/")
        reviewer_model = (settings.scholar_forge_reviewer_model or settings.scholar_forge_organizer_model or "").strip()
        reviewer_provider = (
            settings.scholar_forge_reviewer_provider
            or settings.scholar_forge_organizer_provider
            or settings.default_llm_provider
            or "gemini"
        ).lower()
        return _RoleLLMConfig(
            api_key=reviewer_key,
            base_url=reviewer_url,
            model=reviewer_model,
            provider=reviewer_provider,
        )
    if role == ROLE_IMAGE:
        return _RoleLLMConfig(
            api_key=(settings.scholar_forge_image_api_key or "").strip(),
            base_url=(settings.scholar_forge_image_base_url or "").strip().rstrip("/"),
            model=(settings.scholar_forge_image_model or "").strip(),
            provider=(settings.scholar_forge_image_provider or "qwen").lower(),
        )
    if role == ROLE_POLISHER:
        polisher_key = (settings.scholar_forge_polisher_api_key or settings.scholar_forge_organizer_api_key or "").strip()
        polisher_url = (settings.scholar_forge_polisher_base_url or settings.scholar_forge_organizer_base_url or "").strip().rstrip("/")
        polisher_model = (settings.scholar_forge_polisher_model or settings.scholar_forge_organizer_model or "").strip()
        polisher_provider = (
            settings.scholar_forge_polisher_provider
            or settings.scholar_forge_organizer_provider
            or settings.default_llm_provider
            or "gemini"
        ).lower()
        return _RoleLLMConfig(
            api_key=polisher_key,
            base_url=polisher_url,
            model=polisher_model,
            provider=polisher_provider,
        )
    return _RoleLLMConfig(
        api_key="",
        base_url="",
        model=(settings.default_model or "").strip(),
        provider=(settings.default_llm_provider or "gemini").lower(),
    )


def _role_env_prefix(role: str) -> str:
    if role == ROLE_ORGANIZER:
        return "SCHOLAR_FORGE_ORGANIZER"
    if role == ROLE_WRITER:
        return "SCHOLAR_FORGE_WRITER"
    if role == ROLE_REVIEWER:
        return "SCHOLAR_FORGE_REVIEWER"
    if role == ROLE_IMAGE:
        return "SCHOLAR_FORGE_IMAGE"
    if role == ROLE_POLISHER:
        return "SCHOLAR_FORGE_POLISHER"
    return "SCHOLAR_FORGE"


def _endpoint_label(base_url: str) -> str:
    host = urlparse(base_url).netloc or base_url
    return f"openai-compatible@{host}"


def _timeout_for_role(role: str) -> int:
    timeout = settings.api_timeout
    if role == ROLE_ORGANIZER:
        return max(timeout, 300)
    if role == ROLE_WRITER:
        return max(timeout, 600)
    if role == ROLE_REVIEWER:
        return max(timeout, 300)
    if role == ROLE_POLISHER:
        return max(timeout, 600)
    return timeout


def _generation_params_for_role(role: str) -> Dict[str, Any]:
    return {
        "temperature": 0.35 if role in (ROLE_ORGANIZER, ROLE_REVIEWER) else (0.45 if role == ROLE_POLISHER else 0.65),
        "max_tokens": 16384 if role == ROLE_POLISHER else (8192 if role != ROLE_IMAGE else 2048),
        "timeout": _timeout_for_role(role),
    }


def _create_openai_compatible_caller(api_key: str, base_url: str, model: str, role: str) -> Any:
    from qwen_caller import QwenCaller

    params = _generation_params_for_role(role)
    return QwenCaller(
        api_key=api_key,
        model=model,
        base_url=base_url,
        **params,
    )


def _create_named_provider_caller(provider_str: str, api_key: str, model: str, role: str) -> Any:
    provider = LLMProvider(provider_str)
    if provider not in LLMFactory._callers:
        raise ValueError(f"Unknown LLM provider: {provider_str}")
    caller_class = LLMFactory._callers[provider]
    resolved_model = _model_for_provider(provider_str, model)
    params = _generation_params_for_role(role)
    kwargs: Dict[str, Any] = dict(params)
    if provider_str == "qwen":
        kwargs["base_url"] = settings.qwen_base_url
    return caller_class(api_key=api_key, model=resolved_model, **kwargs)


def _extract_json_block(text: str) -> Optional[dict]:
    if not text:
        return None
    text = text.strip()
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        pass
    match = re.search(r"```(?:json)?\s*([\s\S]*?)```", text)
    if match:
        try:
            return json.loads(match.group(1).strip())
        except json.JSONDecodeError:
            pass
    start = text.find("{")
    end = text.rfind("}")
    if start >= 0 and end > start:
        try:
            return json.loads(text[start : end + 1])
        except json.JSONDecodeError:
            pass
    return None


def _api_key_for_provider(provider_str: str) -> str:
    p = provider_str.lower()
    if p == "gemini":
        return (settings.gemini_api_key or "").strip()
    if p == "qwen":
        return (settings.qwen_api_key or "").strip()
    if p == "mistral":
        return (settings.mistral_api_key or "").strip()
    if p == "groq":
        return (getattr(settings, "groq_api_key", "") or "").strip()
    return ""


def _model_for_provider(provider_str: str, configured_model: str) -> str:
    """Pick a model valid for the resolved provider (fallback to provider default)."""
    m = (configured_model or "").strip()
    p = provider_str.lower()
    if p == "gemini":
        return m if m.startswith("gemini") else settings.gemini_default_model
    if p == "qwen":
        return m if "qwen" in m.lower() else settings.qwen_default_model
    if p == "mistral":
        return m if m.startswith("mistral") else settings.mistral_default_model
    if p == "groq":
        default = getattr(settings, "groq_default_model", "llama-3.3-70b-versatile")
        if not m or m.startswith("gemini") or m.startswith("qwen") or m.startswith("mistral"):
            return default
        return m
    return m or settings.default_model


def _providers_to_try(preferred: str) -> List[str]:
    pref = (preferred or settings.default_llm_provider or "gemini").lower()
    ordered = [pref] + [p for p in _PROVIDER_ORDER if p != pref]
    return [p for p in ordered if _api_key_for_provider(p)]


def _format_api_key_hint(role: str, preferred: str) -> str:
    prefix = _role_env_prefix(role)
    return (
        f"No working LLM config for ScholarForge {role} (configured provider: {preferred}). "
        f"Set {prefix}_API_KEY + {prefix}_BASE_URL + {prefix}_MODEL for a dedicated endpoint "
        f"(e.g. SiliconFlow), or leave {prefix}_API_KEY empty and configure global provider keys "
        f"with {prefix}_PROVIDER / {prefix}_MODEL."
    )


def _is_invalid_api_key_error(exc: Exception) -> bool:
    msg = str(exc).lower()
    return "api_key_invalid" in msg or "api key not found" in msg or "invalid api key" in msg


def _create_llm_for_provider(
    role: str, provider_str: str, configured_model: str, api_key_override: str = ""
) -> Tuple[Any, str, str]:
    api_key = (api_key_override or _api_key_for_provider(provider_str)).strip()
    if not api_key:
        raise ValueError(f"No API key for provider '{provider_str}'")
    model = _model_for_provider(provider_str, configured_model)
    caller = _create_named_provider_caller(provider_str, api_key, model, role)
    return LangChainLLMWrapper(llm_caller=caller), model, provider_str


def _create_llm_from_role_config(role: str) -> Tuple[Any, str, str]:
    cfg = _role_llm_config(role)
    model = cfg.model or settings.default_model
    params = _generation_params_for_role(role)

    if cfg.base_url and not cfg.api_key:
        prefix = _role_env_prefix(role)
        raise ValueError(f"{prefix}_BASE_URL is set but {prefix}_API_KEY is missing")

    if cfg.api_key and cfg.base_url:
        caller = _create_openai_compatible_caller(cfg.api_key, cfg.base_url, model, role)
        label = _endpoint_label(cfg.base_url)
        return LangChainLLMWrapper(llm_caller=caller), model, label

    if cfg.api_key:
        caller = _create_named_provider_caller(cfg.provider, cfg.api_key, model, role)
        resolved_model = _model_for_provider(cfg.provider, model)
        return LangChainLLMWrapper(llm_caller=caller), resolved_model, cfg.provider

    raise ValueError(_format_api_key_hint(role, cfg.provider))


def _has_dedicated_role_credentials(role: str) -> bool:
    cfg = _role_llm_config(role)
    return bool(cfg.api_key or cfg.base_url)


def _resolve_role_llm(role: str) -> Tuple[Any, str, str]:
    """Create an LLM for a ScholarForge role using dedicated or shared credentials."""
    cfg = _role_llm_config(role)
    if _has_dedicated_role_credentials(role):
        return _create_llm_from_role_config(role)

    configured_model = cfg.model
    candidates = _providers_to_try(cfg.provider)
    if not candidates:
        raise ValueError(_format_api_key_hint(role, cfg.provider))

    last_error: Optional[Exception] = None
    for provider_str in candidates:
        try:
            llm, model, prov = _create_llm_for_provider(role, provider_str, configured_model)
            if provider_str != cfg.provider:
                logger.warning(
                    "ScholarForge %s: using %s (%s) — preferred %s unavailable",
                    role,
                    prov,
                    model,
                    cfg.provider,
                )
            return llm, model, prov
        except Exception as e:
            last_error = e
            logger.warning("ScholarForge %s: failed to init %s: %s", role, provider_str, e)

    hint = _format_api_key_hint(role, cfg.provider)
    if last_error:
        raise ValueError(f"{hint} Last error: {last_error}") from last_error
    raise ValueError(hint)


class _LLMJob:
    """Run an LLM call in a background thread; heartbeats keep the HTTP stream alive."""

    def __init__(self, role: str, prompt: str) -> None:
        self.role = role
        self.prompt = prompt
        self.result: Optional[str] = None
        self.error: Optional[Exception] = None
        self.model: Optional[str] = None
        self.provider: Optional[str] = None
        self._thread = threading.Thread(target=self._run, daemon=True)

    def _run(self) -> None:
        cfg = _role_llm_config(self.role)

        if _has_dedicated_role_credentials(self.role):
            try:
                llm, model, prov = _create_llm_from_role_config(self.role)
                self.result = _invoke_llm(llm, self.prompt)
                self.model = model
                self.provider = prov
                return
            except Exception as e:
                if not _is_invalid_api_key_error(e):
                    self.error = e
                    return
                last_error = e
                LLMFactory.clear_cache()
                self.error = ValueError(
                    _format_api_key_hint(self.role, cfg.provider)
                    + f" Last error: {last_error}"
                )
                return

        candidates = _providers_to_try(cfg.provider)
        if not candidates:
            self.error = ValueError(_format_api_key_hint(self.role, cfg.provider))
            return

        last_error: Optional[Exception] = None

        for provider_str in candidates:
            try:
                llm, model, prov = _create_llm_for_provider(
                    self.role, provider_str, cfg.model
                )
                self.result = _invoke_llm(llm, self.prompt)
                self.model = model
                self.provider = prov
                if provider_str != cfg.provider:
                    logger.warning(
                        "ScholarForge %s: fell back to %s (%s) after %s failed",
                        self.role,
                        prov,
                        model,
                        cfg.provider,
                    )
                return
            except Exception as e:
                last_error = e
                if _is_invalid_api_key_error(e):
                    logger.warning(
                        "ScholarForge %s: invalid API key for %s, trying next provider",
                        self.role,
                        provider_str,
                    )
                    LLMFactory.clear_cache()
                    continue
                self.error = e
                return

        self.error = ValueError(
            _format_api_key_hint(self.role, cfg.provider)
            + (f" Last error: {last_error}" if last_error else "")
        )

    def start(self) -> None:
        self._thread.start()

    def join(self, timeout: Optional[float] = None) -> None:
        self._thread.join(timeout=timeout)

    @property
    def alive(self) -> bool:
        return self._thread.is_alive()


def _yield_llm_keepalives(job: _LLMJob, label: str = "Working") -> Iterator[str]:
    """Yield NDJSON heartbeat lines while job runs."""
    elapsed = 0
    while job.alive:
        job.join(timeout=KEEPALIVE_INTERVAL_SEC)
        if job.alive:
            elapsed += KEEPALIVE_INTERVAL_SEC
            yield json.dumps(
                {
                    "type": "heartbeat",
                    "message": f"{label} ({elapsed}s — large prompts can take several minutes)",
                },
                ensure_ascii=False,
            ) + "\n"
    if job.error:
        msg = str(job.error)
        if _is_invalid_api_key_error(job.error):
            prefix = _role_env_prefix(job.role)
            msg += (
                f" — Check {prefix}_API_KEY / {prefix}_BASE_URL / {prefix}_MODEL, "
                f"or set {prefix}_PROVIDER to a provider with a valid global API key."
            )
        yield json.dumps({"type": "error", "message": msg}, ensure_ascii=False) + "\n"
        return
    if job.result is None:
        yield json.dumps({"type": "error", "message": "LLM returned no result"}, ensure_ascii=False) + "\n"


def _doc_type_value(project: ScholarForgeProfile) -> str:
    dt = project.document_type
    return dt.value if hasattr(dt, "value") else str(dt)


def _is_thesis_type(project: ScholarForgeProfile) -> bool:
    return _doc_type_value(project) in ("thesis", "dissertation")


def _format_document_meta_block(project: ScholarForgeProfile) -> str:
    meta = project.document_meta
    if not meta:
        return "Document metadata: (not provided)\n"

    def add(lines: list[str], label: str, value: Optional[str]) -> None:
        if value and str(value).strip():
            lines.append(f"- {label}: {str(value).strip()}")

    lines: list[str] = []
    add(lines, "Author", meta.author_name)
    add(lines, "Email", meta.author_email)
    add(lines, "Affiliation", meta.author_affiliation)
    if meta.co_authors:
        lines.append(f"- Co-authors: {', '.join(meta.co_authors)}")
    add(lines, "University / institution", meta.university)
    add(lines, "Faculty / school", meta.faculty)
    add(lines, "Department", meta.department)
    add(lines, "Degree program", meta.degree_program)
    add(lines, "Degree awarded", meta.degree_awarded)
    add(lines, "Supervisor", meta.supervisor)
    add(lines, "Co-supervisor", meta.co_supervisor)
    add(lines, "Student ID", meta.student_id)
    add(lines, "Submission date", meta.submission_date)
    add(lines, "Location", meta.location)
    add(lines, "Language", meta.language)
    add(lines, "Citation style", meta.citation_style)
    if meta.keywords:
        lines.append(f"- Keywords: {', '.join(meta.keywords)}")
    if meta.abstract_word_limit:
        lines.append(f"- Abstract word limit: {meta.abstract_word_limit}")
    add(lines, "Institutional / formatting requirements", meta.thesis_requirements_notes)

    if not lines:
        return "Document metadata: (not provided)\n"
    return "Document metadata:\n" + "\n".join(lines) + "\n"


def build_front_matter_markdown(project: ScholarForgeProfile) -> str:
    """Title page block included at the start of generated output."""
    meta = project.document_meta
    title = project.structure.document_title if project.structure else project.title
    lines = [f"# {title}", ""]

    if _is_thesis_type(project) and meta:
        if meta.author_name:
            lines.append(f"**Author:** {meta.author_name}")
        if meta.student_id:
            lines.append(f"**Student ID:** {meta.student_id}")
        if meta.degree_program:
            lines.append(f"**Degree program:** {meta.degree_program}")
        if meta.degree_awarded:
            lines.append(f"**Degree:** {meta.degree_awarded}")
        if meta.department:
            lines.append(f"**Department:** {meta.department}")
        if meta.faculty:
            lines.append(f"**Faculty:** {meta.faculty}")
        if meta.university:
            lines.append(f"**University:** {meta.university}")
        if meta.supervisor:
            lines.append(f"**Supervisor:** {meta.supervisor}")
        if meta.co_supervisor:
            lines.append(f"**Co-supervisor:** {meta.co_supervisor}")
        if meta.submission_date:
            lines.append(f"**Submission date:** {meta.submission_date}")
        if meta.location:
            lines.append(f"**Location:** {meta.location}")
        if meta.citation_style:
            lines.append(f"**Citation style:** {meta.citation_style}")
        if meta.keywords:
            lines.append(f"**Keywords:** {', '.join(meta.keywords)}")
    elif meta:
        if meta.author_name:
            lines.append(f"**Author:** {meta.author_name}")
        if meta.author_affiliation:
            lines.append(f"**Affiliation:** {meta.author_affiliation}")
        if meta.co_authors:
            lines.append(f"**Co-authors:** {', '.join(meta.co_authors)}")
        if meta.citation_style:
            lines.append(f"**Citation style:** {meta.citation_style}")
        if meta.keywords:
            lines.append(f"**Keywords:** {', '.join(meta.keywords)}")

    return "\n".join(lines) + "\n\n---\n\n"


def _resolve_role(role: str) -> Tuple[Any, str, str]:
    return _resolve_role_llm(role)


def _invoke_llm(llm: Any, prompt: str) -> str:
    result = llm.invoke(prompt)
    return result if isinstance(result, str) else str(result)


def _project_context_block(project: ScholarForgeProfile) -> str:
    sites = "\n".join(f"- {s}" for s in (project.recommended_sites or [])) or "(none)"
    answers = json.dumps(project.clarification_answers or {}, ensure_ascii=False, indent=2)
    return (
        f"Subject: {project.subject}\n"
        f"Title: {project.title}\n"
        f"Document type: {_doc_type_value(project)}\n"
        f"{_format_document_meta_block(project)}\n"
        f"Short intro: {project.short_intro}\n"
        f"Detailed requirements:\n{project.detailed_prompt}\n\n"
        f"Recommended reference sites:\n{sites}\n\n"
        f"User clarification answers:\n{answers}\n"
    )


def _images_dir(project_id: str) -> str:
    path = os.path.join(settings.data_directory, "scholar_forge_images", project_id)
    os.makedirs(path, exist_ok=True)
    return path


def _openai_vision_chat(
    api_key: str,
    base_url: str,
    model: str,
    image_bytes: bytes,
    prompt: str,
    *,
    timeout: int = 90,
) -> str:
    b64 = base64.b64encode(image_bytes).decode("utf-8")
    data_url = f"data:image/jpeg;base64,{b64}"
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json",
    }
    payload = {
        "model": model,
        "messages": [
            {
                "role": "user",
                "content": [
                    {"type": "image_url", "image_url": {"url": data_url}},
                    {"type": "text", "text": prompt},
                ],
            }
        ],
    }
    resp = requests.post(
        f"{base_url.rstrip('/')}/chat/completions",
        headers=headers,
        json=payload,
        timeout=timeout,
    )
    if resp.status_code == 200:
        data = resp.json()
        choices = data.get("choices") or []
        if choices:
            return choices[0].get("message", {}).get("content", "") or ""
    raise RuntimeError(f"Vision API failed ({resp.status_code}): {resp.text[:300]}")


def describe_image_for_academic(image_bytes: bytes, filename: str, subject: str) -> str:
    """Use configured vision model to describe an academic figure."""
    cfg = _role_llm_config(ROLE_IMAGE)
    model = cfg.model or settings.scholar_forge_image_model

    prompt = (
        f"You are an academic figure analyst for the subject: {subject}.\n"
        "Describe this image in detail for placement in a scholarly article or thesis.\n"
        "Include: (1) what the figure shows, (2) key labels/data visible, "
        "(3) suggested section where it fits (e.g. Methods, Results), "
        "(4) a concise caption suitable for publication.\n"
        "Respond in plain text, structured with short headings."
    )

    if cfg.api_key and cfg.base_url:
        try:
            return _openai_vision_chat(cfg.api_key, cfg.base_url, model, image_bytes, prompt)
        except Exception as e:
            logger.warning("ScholarForge image vision (%s) failed for %s: %s", cfg.base_url, filename, e)

    vision_key = cfg.api_key or settings.qwen_api_key
    vision_base = cfg.base_url or settings.qwen_base_url
    if vision_key and (cfg.provider == "qwen" or cfg.base_url or cfg.api_key):
        try:
            return _openai_vision_chat(vision_key, vision_base, model, image_bytes, prompt)
        except Exception as e:
            logger.warning("Qwen vision failed for %s: %s", filename, e)

    llm, _, _ = _resolve_role(ROLE_ORGANIZER)
    return _invoke_llm(
        llm,
        f"{prompt}\n\n(Image file: {filename}. Vision unavailable — infer typical figure needs from filename.)",
    )


def _build_clarification_html(form: ScholarForgeClarificationForm) -> str:
    rows = []
    for field in form.fields:
        req = "required" if field.required else ""
        if field.field_type == "textarea":
            rows.append(
                f'<label for="{field.id}">{field.label}</label>'
                f'<textarea id="{field.id}" name="{field.id}" rows="4" {req} '
                f'placeholder="{field.placeholder or ""}"></textarea>'
            )
        elif field.field_type == "select" and field.options:
            opts = "".join(f'<option value="{o}">{o}</option>' for o in field.options)
            rows.append(
                f'<label for="{field.id}">{field.label}</label>'
                f'<select id="{field.id}" name="{field.id}" {req}>{opts}</select>'
            )
        else:
            rows.append(
                f'<label for="{field.id}">{field.label}</label>'
                f'<input type="text" id="{field.id}" name="{field.id}" {req} '
                f'placeholder="{field.placeholder or ""}" />'
            )
        if field.help_text:
            rows.append(f'<p class="help">{field.help_text}</p>')
    body = "\n".join(rows)
    return (
        f'<!DOCTYPE html><html><head><meta charset="utf-8">'
        f"<title>{form.title}</title>"
        f"<style>body{{font-family:Georgia,serif;max-width:720px;margin:2rem auto;"
        f"padding:1rem}}label{{display:block;margin-top:1rem;font-weight:bold}}"
        f"input,textarea,select{{width:100%;padding:0.5rem;margin-top:0.25rem}}"
        f".help{{font-size:0.9rem;color:#555}}</style></head>"
        f"<body><h1>{form.title}</h1><p>{form.intro}</p><form>{body}</form></body></html>"
    )


def stream_prepare_session(
    project: ScholarForgeProfile,
    rag_system: Any,
    pdf_reader: Any,
) -> Iterator[str]:
    yield json.dumps({"type": "log", "message": "Preparing ScholarForge session..."}, ensure_ascii=False) + "\n"

    processed_materials = list(project.materials or [])
    for mat in processed_materials:
        if mat.format.lower() == "pdf" and mat.content.startswith("base64:"):
            try:
                raw = base64.b64decode(mat.content[7:])
                extracted = pdf_reader._extract_text_from_pdf(raw)
                if extracted.get("success"):
                    mat.content = extracted["text"]
                    mat.format = "pdf_text"
                    yield json.dumps(
                        {"type": "log", "message": f"Extracted PDF text from {mat.filename}."},
                        ensure_ascii=False,
                    ) + "\n"
            except Exception as e:
                yield json.dumps(
                    {"type": "log", "message": f"PDF extraction warning ({mat.filename}): {e}"},
                    ensure_ascii=False,
                ) + "\n"

    project.materials = processed_materials
    collection = ingest_materials_to_rag(rag_system, project, processed_materials)
    project.rag_collection = collection
    project.status = ScholarForgeStatus.PREPARING

    # Drop bulky file payloads from persisted profile — content now lives in dedicated RAG.
    project.materials = [
        ScholarForgeMaterialInput(
            filename=m.filename,
            format=m.format,
            content="[ingested]",
            description=m.description,
        )
        for m in processed_materials
    ]

    img_dir = _images_dir(project.id)
    for img in project.images or []:
        if img.stored_filename:
            fpath = os.path.join(img_dir, img.stored_filename)
            if os.path.isfile(fpath) and not img.description:
                with open(fpath, "rb") as f:
                    desc = describe_image_for_academic(f.read(), img.filename, project.subject)
                img.description = desc
                cap_match = re.search(r"(?i)caption[:\s]+(.+)", desc)
                img.caption = cap_match.group(1).strip() if cap_match else img.filename
                hint_match = re.search(r"(?i)(?:section|placement)[:\s]+(.+)", desc)
                img.placement_hint = hint_match.group(1).strip()[:200] if hint_match else None
                yield json.dumps(
                    {"type": "log", "message": f"Described image: {img.filename}"},
                    ensure_ascii=False,
                ) + "\n"

    project.status = ScholarForgeStatus.CLARIFYING
    yield json.dumps(
        {
            "type": "session_ready",
            "rag_collection": collection,
            "status": project.status.value,
        },
        ensure_ascii=False,
    ) + "\n"


def stream_clarification(project: ScholarForgeProfile, rag_system: Any) -> Iterator[str]:
    yield json.dumps({"type": "log", "message": "Organizer analyzing requirements..."}, ensure_ascii=False) + "\n"

    try:
        rag_ctx = query_project_rag(
            rag_system,
            project.rag_collection or "",
            f"{project.subject} {project.title} requirements scope methodology",
            RAG_QUERY_RESULTS,
        )
        images_block = ""
        if project.images:
            images_block = "\n\nUploaded figures:\n" + "\n".join(
                f"- {i.filename}: {(i.description or '')[:300]}" for i in project.images
            )

        if settings.scholar_forge_use_enhanced_prompts:
            prompt = enhanced_clarification_prompt(project, rag_ctx, images_block)
        else:
            prompt = (
                "You are the ScholarForge organizer AI (MCP controller) for academic writing.\n"
                "Analyze whether the user's requirements are sufficient to plan and write a "
                f"{_doc_type_value(project)}.\n\n"
                + _project_context_block(project)
                + (f"\n\nRAG context from uploaded materials:\n{rag_ctx[:8000]}" if rag_ctx else "")
                + images_block
                + "\n\nRespond with ONLY valid JSON:\n"
                "{\n"
                '  "sufficient": true|false,\n'
                '  "title": "form title if insufficient",\n'
                '  "intro": "why more info is needed",\n'
                '  "fields": [\n'
                '    {"id":"field_id","label":"Question","field_type":"text|textarea|select|number",'
                '"required":true,"help_text":"...","options":[],"placeholder":"..."}\n'
                "  ]\n"
                "}\n"
                "If sufficient=true, fields may be empty. Ask 3-8 targeted academic questions when insufficient."
            )

        yield json.dumps(
            {"type": "log", "message": "Calling organizer model (streaming keepalive active)..."},
            ensure_ascii=False,
        ) + "\n"

        job = _LLMJob(ROLE_ORGANIZER, prompt)
        job.start()
        yield from _yield_llm_keepalives(job, "Organizer analyzing requirements")
        if job.error:
            return
        raw = job.result or ""
        model = job.model or settings.scholar_forge_organizer_model
        provider = job.provider or settings.scholar_forge_organizer_provider

        data = _extract_json_block(raw) or {"sufficient": True, "fields": []}

        fields = []
        for f in data.get("fields") or []:
            try:
                fields.append(ScholarForgeFormField(**f))
            except Exception:
                continue

        form = ScholarForgeClarificationForm(
            title=data.get("title") or "Additional academic details required",
            intro=data.get("intro") or "",
            fields=fields,
            sufficient=bool(data.get("sufficient")),
        )
        if not form.sufficient and form.fields:
            form.html_preview = _build_clarification_html(form)

        project.clarification = form
        project.status = (
            ScholarForgeStatus.STRUCTURING if form.sufficient else ScholarForgeStatus.CLARIFYING
        )

        yield json.dumps(
            {
                "type": "clarification",
                "sufficient": form.sufficient,
                "form": form.model_dump(),
                "model_used": model,
                "provider": provider,
            },
            ensure_ascii=False,
        ) + "\n"
    except Exception as e:
        logger.exception("ScholarForge clarify failed")
        yield json.dumps({"type": "error", "message": str(e)}, ensure_ascii=False) + "\n"


def stream_generate_structure(project: ScholarForgeProfile, rag_system: Any) -> Iterator[str]:
    yield json.dumps({"type": "log", "message": "Generating document structure..."}, ensure_ascii=False) + "\n"

    try:
        rag_ctx = query_project_rag(
            rag_system,
            project.rag_collection or "",
            f"{project.subject} outline structure chapters sections",
            RAG_QUERY_RESULTS,
        )

        type_hint = {
            "article": "journal-style article with abstract, introduction, methods, results, discussion, conclusion, references",
            "thesis": (
                "master's thesis with standard front matter (title page, declaration/ethics statement, "
                "abstract, acknowledgements, table of contents, list of figures/tables), numbered chapters, "
                "conclusion, bibliography, and appendices if needed — follow the document metadata and "
                "institutional requirements provided"
            ),
            "dissertation": (
                "doctoral dissertation with comprehensive front matter, numbered chapters, "
                "bibliography, and appendices — follow the document metadata and institutional requirements"
            ),
        }.get(_doc_type_value(project), "academic document")

        if settings.scholar_forge_use_enhanced_prompts:
            prompt = enhanced_structure_prompt(project, rag_ctx)
        else:
            prompt = (
            f"You are the ScholarForge organizer. Create a detailed structure plan for a {type_hint}.\n\n"
            + _project_context_block(project)
            + (f"\n\nMaterial context:\n{rag_ctx[:6000]}" if rag_ctx else "")
            + "\n\nRespond ONLY with JSON:\n"
            "{\n"
            f'  "document_title": "{project.title}",\n'
            '  "abstract_outline": "brief abstract plan",\n'
            '  "notes": "guidance for writer",\n'
            '  "sections": [\n'
            '    {"id":"sec_1","title":"Section title","description":"what to cover",'
            '"order":1,"word_target":1500}\n'
            "  ]\n"
            "}\n"
            "Include all standard academic sections appropriate for the document type. "
            "Honor citation style, language, and any institutional formatting notes in document metadata."
        )

        job = _LLMJob(ROLE_ORGANIZER, prompt)
        job.start()
        yield from _yield_llm_keepalives(job, "Building structure plan")
        if job.error:
            return
        raw = job.result or ""
        model = job.model or settings.scholar_forge_organizer_model
        provider = job.provider or settings.scholar_forge_organizer_provider
        data = _extract_json_block(raw) or {}
        sections = []
        for i, s in enumerate(data.get("sections") or []):
            sid = s.get("id") or f"sec_{i+1}"
            sections.append(
                ScholarForgeSection(
                    id=sid,
                    title=s.get("title") or f"Section {i+1}",
                    description=s.get("description") or "",
                    order=int(s.get("order") or i + 1),
                    word_target=s.get("word_target"),
                )
            )
        structure = ScholarForgeStructure(
            document_title=data.get("document_title") or project.title,
            abstract_outline=data.get("abstract_outline") or "",
            sections=sorted(sections, key=lambda x: x.order),
            notes=data.get("notes") or "",
        )
        project.structure = structure
        project.status = ScholarForgeStatus.STRUCTURING

        yield json.dumps(
            {
                "type": "structure",
                "structure": structure.model_dump(),
                "model_used": model,
                "provider": provider,
            },
            ensure_ascii=False,
        ) + "\n"
    except Exception as e:
        logger.exception("ScholarForge structure failed")
        yield json.dumps({"type": "error", "message": str(e)}, ensure_ascii=False) + "\n"


def stream_confirm_and_plan(project: ScholarForgeProfile, rag_system: Any, notes: str = "") -> Iterator[str]:
    if not project.structure:
        yield json.dumps({"type": "error", "message": "No structure to confirm."}, ensure_ascii=False) + "\n"
        return

    sections_text = "\n".join(
        f"{s.order}. {s.title}: {s.description}" for s in project.structure.sections
    )
    prompt = (
        "You are the ScholarForge organizer. Convert this approved structure into a "
        "detailed writing plan the writer AI will follow section-by-section.\n\n"
        + _project_context_block(project)
        + f"\n\nApproved structure:\n{sections_text}\n"
        f"\nUser notes: {notes or '(none)'}\n"
        "\nOutput a clear markdown plan with per-section bullet instructions, "
        "citation expectations, tone, and cross-references."
    )
    try:
        job = _LLMJob(ROLE_ORGANIZER, prompt)
        job.start()
        yield from _yield_llm_keepalives(job, "Building writing plan")
        if job.error:
            return
        plan = job.result or ""
        model = job.model or settings.scholar_forge_organizer_model
        provider = job.provider or settings.scholar_forge_organizer_provider
        project.final_plan = plan
        project.status = ScholarForgeStatus.PLANNING

        yield json.dumps(
            {
                "type": "plan",
                "final_plan": plan,
                "model_used": model,
                "provider": provider,
            },
            ensure_ascii=False,
        ) + "\n"
    except Exception as e:
        logger.exception("ScholarForge confirm plan failed")
        yield json.dumps({"type": "error", "message": str(e)}, ensure_ascii=False) + "\n"


def _paragraph_count_for_section(section: ScholarForgeSection) -> int:
    target = section.word_target or 1200
    return max(3, min(12, target // 150))


def _parse_review_report(text: str) -> ScholarForgeReviewReport:
    data = _extract_json_block(text) or {}
    score = int(data.get("quality_score") or data.get("score") or 0)
    approved = bool(data.get("approved", False))
    if not approved and score >= settings.scholar_forge_review_pass_score:
        approved = True
    issues = data.get("issues") or []
    suggestions = data.get("suggestions") or data.get("recommendations") or []
    if isinstance(issues, str):
        issues = [issues]
    if isinstance(suggestions, str):
        suggestions = [suggestions]
    return ScholarForgeReviewReport(
        approved=approved,
        quality_score=max(0, min(100, score)),
        issues=[str(i) for i in issues if i],
        suggestions=[str(s) for s in suggestions if s],
        summary=str(data.get("summary") or data.get("report") or "")[:2000],
    )


def _emit_pipeline_step(
    section: ScholarForgeSection,
    step_id: str,
    step_type: str,
    paragraph_index: Optional[int],
    status: str,
    label: str,
    detail: Optional[str] = None,
) -> ScholarForgePipelineStep:
    step = ScholarForgePipelineStep(
        step_id=step_id,
        step_type=step_type,
        section_id=section.id,
        paragraph_index=paragraph_index,
        status=status,
        label=label,
        detail=detail,
    )
    section.pipeline_steps = (section.pipeline_steps or []) + [step]
    return step


def _write_paragraph_prompt(
    project: ScholarForgeProfile,
    section: ScholarForgeSection,
    para_idx: int,
    para_total: int,
    prior_in_section: str,
    prior_sections: str,
    rag_ctx: str,
    images_catalog: str,
    reference_context: str = "",
) -> str:
    """Build the writer prompt for a single paragraph. Uses enhanced prompt when enabled."""
    if settings.scholar_forge_use_enhanced_prompts:
        return enhanced_writer_paragraph_prompt(
            project=project,
            section_title=section.title,
            section_description=section.description,
            section_order=section.order,
            word_target=section.word_target or 1200,
            para_idx=para_idx,
            para_total=para_total,
            prior_in_section=prior_in_section,
            prior_sections=prior_sections,
            rag_ctx=rag_ctx,
            reference_context=reference_context,
            images_catalog=images_catalog,
        )
    return (
        f"You are an expert academic writer composing a {_doc_type_value(project)}.\n\n"
        f"## Overall plan\n{project.final_plan[:12000]}\n\n"
        f"## Current section ({section.order}): {section.title}\n"
        f"Section brief: {section.description}\n"
        f"Target section length: ~{section.word_target or 1200} words total\n\n"
        f"## Project context\n{_project_context_block(project)[:8000]}\n\n"
        + (f"## Retrieved materials\n{rag_ctx[:8000]}\n\n" if rag_ctx else "")
        + (f"## Prior sections summary\n{prior_sections}\n\n" if prior_sections else "")
        + (f"## Paragraphs already written in this section\n{prior_in_section}\n\n" if prior_in_section else "")
        + (f"## Gathered References\n{reference_context[:6000]}\n\n" if reference_context else "")
        + images_catalog
        + f"\n\nWrite ONLY paragraph {para_idx + 1} of approximately {para_total} for this section. "
        f"Write 1 cohesive scholarly paragraph (120–180 words). Do not repeat prior paragraphs. "
        f"Use {project.document_meta.citation_style or 'APA'} citations where appropriate. "
        "Return ONLY the paragraph text — no headings, labels, or commentary."
    )


def _review_paragraph_prompt(
    project: ScholarForgeProfile,
    section: ScholarForgeSection,
    paragraph_text: str,
    para_idx: int,
) -> str:
    """Build the reviewer critique prompt. Uses enhanced prompt when enabled."""
    if settings.scholar_forge_use_enhanced_prompts:
        return enhanced_reviewer_prompt(
            project=project,
            section_title=section.title,
            paragraph_text=paragraph_text,
            para_idx=para_idx,
        )
    return (
        "You are the ScholarForge reviewer agent. Critique this paragraph for academic quality, "
        "coherence with the section plan, clarity, evidence, and terminology. "
        "Do NOT rewrite the paragraph — return structured feedback only.\n\n"
        f"Document type: {_doc_type_value(project)}\n"
        f"Section: {section.title}\n"
        f"Plan excerpt:\n{(project.final_plan or '')[:4000]}\n\n"
        f"Paragraph {para_idx + 1} draft:\n{paragraph_text}\n\n"
        "Respond with JSON only:\n"
        "{\n"
        '  "approved": true/false,\n'
        '  "quality_score": 0-100,\n'
        '  "issues": ["issue1", "issue2"],\n'
        '  "suggestions": ["specific edit 1", "specific edit 2"],\n'
        '  "summary": "one-paragraph review summary"\n'
        "}\n"
        f"Set approved=true if quality_score >= {settings.scholar_forge_review_pass_score} "
        "and no major issues remain."
    )


def _revise_paragraph_prompt(
    project: ScholarForgeProfile,
    section: ScholarForgeSection,
    paragraph_text: str,
    review: ScholarForgeReviewReport,
    para_idx: int,
) -> str:
    """Build the paragraph revision prompt. Uses enhanced prompt when enabled."""
    issues = "\n".join(f"- {i}" for i in review.issues) or "(none listed)"
    suggestions = "\n".join(f"- {s}" for s in review.suggestions) or "(none listed)"
    if settings.scholar_forge_use_enhanced_prompts:
        return enhanced_revise_prompt(
            project=project,
            section_title=section.title,
            paragraph_text=paragraph_text,
            review_summary=review.summary,
            issues=review.issues,
            suggestions=review.suggestions,
            para_idx=para_idx,
        )
    return (
        f"You are the ScholarForge writer. Revise this paragraph based on the reviewer report.\n\n"
        f"Section: {section.title}\n"
        f"Paragraph {para_idx + 1} current draft:\n{paragraph_text}\n\n"
        f"## Reviewer summary\n{review.summary}\n\n"
        f"## Issues\n{issues}\n\n"
        f"## Suggested changes\n{suggestions}\n\n"
        "Apply the feedback while preserving factual content and citation style. "
        "Return ONLY the revised paragraph — no commentary."
    )



def _run_paragraph_pipeline(
    project: ScholarForgeProfile,
    section: ScholarForgeSection,
    para_idx: int,
    para_total: int,
    prior_in_section: str,
    prior_sections: str,
    rag_ctx: str,
    images_catalog: str,
    writer_llm: Any,
    reference_context: str = "",
) -> Iterator[str]:
    """Write → review → revise loop for one paragraph. Yields NDJSON events."""
    record = ScholarForgeParagraphRecord(index=para_idx, status="writing")
    section.paragraphs = (section.paragraphs or []) + [record]

    step_write = _emit_pipeline_step(
        section, f"{section.id}-p{para_idx}-write", "write", para_idx, "running",
        f"Writer: paragraph {para_idx + 1}/{para_total}",
    )
    yield json.dumps(
        {
            "type": "flow_step",
            "step": step_write.model_dump(),
            "section_id": section.id,
            "paragraph_index": para_idx,
        },
        ensure_ascii=False,
    ) + "\n"
    yield json.dumps(
        {
            "type": "paragraph_start",
            "section_id": section.id,
            "paragraph_index": para_idx,
            "total": para_total,
        },
        ensure_ascii=False,
    ) + "\n"

    write_prompt = _write_paragraph_prompt(
        project, section, para_idx, para_total, prior_in_section, prior_sections, rag_ctx, images_catalog,
        reference_context=reference_context,
    )
    paragraph_text = ""
    stream_iter = getattr(writer_llm, "stream", None)
    if stream_iter:
        for chunk in stream_iter(write_prompt):
            piece = chunk if isinstance(chunk, str) else str(chunk)
            if piece:
                paragraph_text += piece
                yield json.dumps(
                    {"type": "paragraph_delta", "section_id": section.id, "paragraph_index": para_idx, "text": piece},
                    ensure_ascii=False,
                ) + "\n"
    if not paragraph_text.strip():
        job = _LLMJob(ROLE_WRITER, write_prompt)
        job.start()
        yield from _yield_llm_keepalives(job, f"Writing paragraph {para_idx + 1}")
        if job.error:
            step_write.status = "failed"
            return
        paragraph_text = job.result or ""

    record.draft = paragraph_text
    step_write.status = "done"

    max_rounds = max(1, settings.scholar_forge_max_review_rounds)
    for review_round in range(max_rounds):
        record.status = "reviewing"
        step_review = _emit_pipeline_step(
            section,
            f"{section.id}-p{para_idx}-review-{review_round}",
            "review",
            para_idx,
            "running",
            f"Reviewer: paragraph {para_idx + 1} (round {review_round + 1})",
        )
        yield json.dumps(
            {"type": "flow_step", "step": step_review.model_dump(), "section_id": section.id, "paragraph_index": para_idx},
            ensure_ascii=False,
        ) + "\n"
        yield json.dumps(
            {"type": "review_start", "section_id": section.id, "paragraph_index": para_idx, "round": review_round + 1},
            ensure_ascii=False,
        ) + "\n"

        review_prompt = _review_paragraph_prompt(project, section, paragraph_text, para_idx)
        review_job = _LLMJob(ROLE_REVIEWER, review_prompt)
        review_job.start()
        yield from _yield_llm_keepalives(review_job, f"Reviewing paragraph {para_idx + 1}")
        if review_job.error:
            step_review.status = "failed"
            return
        report = _parse_review_report(review_job.result or "")
        record.reviews = (record.reviews or []) + [report]
        step_review.status = "done"

        yield json.dumps(
            {
                "type": "review_report",
                "section_id": section.id,
                "paragraph_index": para_idx,
                "round": review_round + 1,
                "report": report.model_dump(),
            },
            ensure_ascii=False,
        ) + "\n"

        if report.approved:
            record.status = "approved"
            break

        record.status = "revising"
        record.revision_rounds += 1
        step_revise = _emit_pipeline_step(
            section,
            f"{section.id}-p{para_idx}-revise-{review_round}",
            "revise",
            para_idx,
            "running",
            f"Writer revises paragraph {para_idx + 1} (round {review_round + 1})",
        )
        yield json.dumps(
            {"type": "flow_step", "step": step_revise.model_dump(), "section_id": section.id, "paragraph_index": para_idx},
            ensure_ascii=False,
        ) + "\n"
        yield json.dumps(
            {"type": "revision_start", "section_id": section.id, "paragraph_index": para_idx, "round": review_round + 1},
            ensure_ascii=False,
        ) + "\n"

        revise_prompt = _revise_paragraph_prompt(project, section, paragraph_text, report, para_idx)
        revised = ""
        if stream_iter:
            for chunk in stream_iter(revise_prompt):
                piece = chunk if isinstance(chunk, str) else str(chunk)
                if piece:
                    revised += piece
                    yield json.dumps(
                        {"type": "revision_delta", "section_id": section.id, "paragraph_index": para_idx, "text": piece},
                        ensure_ascii=False,
                    ) + "\n"
        if not revised.strip():
            revise_job = _LLMJob(ROLE_WRITER, revise_prompt)
            revise_job.start()
            yield from _yield_llm_keepalives(revise_job, f"Revising paragraph {para_idx + 1}")
            if revise_job.error:
                step_revise.status = "failed"
                return
            revised = revise_job.result or paragraph_text
        paragraph_text = revised.strip() or paragraph_text
        step_revise.status = "done"

    record.content = paragraph_text.strip()
    record.status = "approved"
    yield json.dumps(
        {
            "type": "paragraph_done",
            "section_id": section.id,
            "paragraph_index": para_idx,
            "text": record.content,
            "revision_rounds": record.revision_rounds,
            "review_count": len(record.reviews),
        },
        ensure_ascii=False,
    ) + "\n"


def stream_generate_document(project: ScholarForgeProfile, rag_system: Any) -> Iterator[str]:
    if not project.structure or not project.final_plan:
        yield json.dumps({"type": "error", "message": "Confirm structure and plan first."}, ensure_ascii=False) + "\n"
        return

    try:
        writer_llm, writer_model, writer_provider = _resolve_role(ROLE_WRITER)
    except Exception as e:
        yield json.dumps({"type": "error", "message": str(e)}, ensure_ascii=False) + "\n"
        return

    project.status = ScholarForgeStatus.GENERATING
    sections = sorted(project.structure.sections, key=lambda s: s.order)
    accumulated_summaries: List[str] = []
    full_parts: List[str] = [build_front_matter_markdown(project)]

    if project.structure.abstract_outline:
        full_parts.append(f"## Abstract\n\n{project.structure.abstract_outline}\n")

    # ── Enhanced: gather references via web search ──
    reference_context = ""
    if settings.scholar_forge_web_search_enabled:
        try:
            from .web_search_service import web_search
            search_func = web_search
        except Exception:
            search_func = None

        if search_func:
            yield json.dumps(
                {"type": "log", "message": "Gathering academic references from web search..."},
                ensure_ascii=False,
            ) + "\n"
            try:
                from .scholar_forge_enhanced import gather_references as _gather_refs
                reference_context, _bib = _gather_refs(project, search_func)
                if reference_context:
                    yield json.dumps(
                        {"type": "log", "message": f"Found references for writing context."},
                        ensure_ascii=False,
                    ) + "\n"
            except Exception as e:
                logger.warning("ScholarForge reference gathering failed: %s", e)

    images_catalog = ""
    if project.images:
        images_catalog = "\n\nAvailable figures for placement:\n" + "\n".join(
            f"- [{i.filename}] {i.caption or i.filename}: {(i.description or '')[:400]}"
            for i in project.images
        )

    for idx, section in enumerate(sections):
        section.paragraphs = []
        section.pipeline_steps = []
        section.status = "generating"
        yield json.dumps(
            {"type": "section_start", "section_id": section.id, "title": section.title, "index": idx + 1},
            ensure_ascii=False,
        ) + "\n"

        rag_ctx = query_project_rag(
            rag_system,
            project.rag_collection or "",
            f"{section.title} {section.description} {project.subject}",
            RAG_QUERY_RESULTS,
        )
        prior_sections = "\n".join(accumulated_summaries[-5:])
        para_total = _paragraph_count_for_section(section)
        section_paragraphs: List[str] = []

        for para_idx in range(para_total):
            prior_in_section = "\n\n".join(section_paragraphs)
            pipeline = _run_paragraph_pipeline(
                project,
                section,
                para_idx,
                para_total,
                prior_in_section,
                prior_sections,
                rag_ctx,
                images_catalog,
                writer_llm,
                reference_context=reference_context,
            )
            for line in pipeline:
                yield line
            # A paragraph is only complete when its record has approved content.
            last = section.paragraphs[-1] if section.paragraphs else None
            if not last or last.status != "approved" or not last.content:
                # The pipeline already emitted an error line before returning.
                return
            section_paragraphs.append(last.content)

        polished = "\n\n".join(p for p in section_paragraphs if p.strip())
        if not polished.strip():
            yield json.dumps(
                {"type": "error", "message": f"No content generated for section {section.title}"},
                ensure_ascii=False,
            ) + "\n"
            return

        yield json.dumps({"type": "log", "message": f"Summarizing section: {section.title}"}, ensure_ascii=False) + "\n"
        summary_prompt = f"Summarize this section in 3-5 bullet points for continuity:\n\n{polished[:6000]}"
        summary_job = _LLMJob(ROLE_ORGANIZER, summary_prompt)
        summary_job.start()
        yield from _yield_llm_keepalives(summary_job, "Summarizing section")
        if summary_job.error:
            return
        summary = summary_job.result or ""

        section.content = polished
        section.summary = summary
        section.status = "reviewed"
        project.section_cache[section.id] = polished
        accumulated_summaries.append(f"**{section.title}**: {summary[:500]}")

        full_parts.append(f"## {section.title}\n\n{polished.strip()}\n")
        yield json.dumps(
            {
                "type": "section_done",
                "section_id": section.id,
                "summary": summary[:500],
                "paragraph_count": len(section_paragraphs),
            },
            ensure_ascii=False,
        ) + "\n"

    if project.images:
        full_parts.append("\n## Figures\n")
        for img in project.images:
            cap = img.caption or img.filename
            full_parts.append(f"**Figure — {img.filename}**: {cap}\n")

    full_parts.append("\n## References\n\n*(References to be finalized from in-text citations.)*\n")
    body = "\n".join(full_parts)

    # ── Enhanced: post-generation polishing pass ──
    if settings.scholar_forge_polish_enabled:
        try:
            polisher_cfg = _role_llm_config(ROLE_POLISHER)
            polisher_model = polisher_cfg.model or settings.scholar_forge_polisher_model
            polisher_provider = polisher_cfg.provider or settings.scholar_forge_polisher_provider

            polish_prompt = build_polisher_prompt(
                project=project,
                full_body=body,
                reference_context=reference_context,
            )

            yield json.dumps(
                {"type": "log", "message": "Running post-generation polish (transitions, citations, formatting)..."},
                ensure_ascii=False,
            ) + "\n"

            polish_job = _LLMJob(ROLE_POLISHER, polish_prompt)
            polish_job.start()
            yield from _yield_llm_keepalives(polish_job, "Polishing document")
            if polish_job.error:
                yield json.dumps(
                    {"type": "log", "message": f"Polishing skipped: {polish_job.error}. Unpolished version saved."},
                    ensure_ascii=False,
                ) + "\n"
            elif polish_job.result:
                polished = extract_polished_document(polish_job.result)
                if polished and len(polished) > len(body) * 0.3:
                    body = polished
                    yield json.dumps(
                        {"type": "log", "message": "Document polished — transitions, citations, and formatting refined."},
                        ensure_ascii=False,
                    ) + "\n"
        except Exception as e:
            logger.warning("ScholarForge polishing failed: %s", e)
            yield json.dumps(
                {"type": "log", "message": f"Polishing unavailable: {e}. Unpolished version saved."},
                ensure_ascii=False,
            ) + "\n"

    md_id, md_name = save_markdown(body, project.title)
    project.output_markdown_id = md_id

    pdf_id: Optional[str] = None
    pdf_name: Optional[str] = None
    try:
        pdf_id, pdf_name = markdown_to_pdf(body, project.title)
        project.output_pdf_id = pdf_id
    except Exception as e:
        logger.exception("ScholarForge PDF export failed (markdown still saved): %s", e)
        yield json.dumps(
            {
                "type": "log",
                "message": f"PDF export skipped: {e}. Markdown download is still available.",
            },
            ensure_ascii=False,
        ) + "\n"

    project.status = ScholarForgeStatus.COMPLETED

    yield json.dumps(
        {
            "type": "done",
            "markdown_id": md_id,
            "markdown_filename": md_name,
            "pdf_id": pdf_id,
            "pdf_filename": pdf_name,
            "writer_model": writer_model,
            "writer_provider": writer_provider,
            "reviewer_model": settings.scholar_forge_reviewer_model or settings.scholar_forge_organizer_model,
            "organizer_model": settings.scholar_forge_organizer_model,
            "pipeline": "write-review-revise",
        },
        ensure_ascii=False,
    ) + "\n"


def store_uploaded_image(project_id: str, filename: str, data: bytes) -> str:
    stored = f"{uuid.uuid4().hex[:12]}_{filename.replace('/', '_')}"
    path = os.path.join(_images_dir(project_id), stored)
    with open(path, "wb") as f:
        f.write(data)
    return stored
