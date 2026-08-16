"""
Multi-chapter article generation with NDJSON streaming for the Articles module.
"""
from __future__ import annotations

import json
import logging
import os
import uuid
from typing import Any, Iterator, List, Optional

from .config import settings
from .llm_factory import LLMFactory, LLMProvider
from .llm_langchain_wrapper import LangChainLLMWrapper
from .models import ArticleProfile, AssistantProfile, LLMProviderType

logger = logging.getLogger(__name__)

SOURCE_MAX_CHARS = 120_000
PREV_MAX_CHARS = 24_000


def _truncate(text: str, max_len: int) -> str:
    t = (text or "").strip()
    if len(t) <= max_len:
        return t
    return t[: max_len - 20] + "\n\n[... truncated ...]"


def _resolve_llm(article: ArticleProfile) -> Tuple[Any, str, LLMProviderType]:
    provider_str = (
        article.llm_provider.value
        if article.llm_provider
        else settings.default_llm_provider
    )
    if provider_str == "gemini":
        provider = LLMProviderType.GEMINI
        api_key = settings.gemini_api_key
        model_name = article.model_name or settings.gemini_default_model
    elif provider_str == "qwen":
        provider = LLMProviderType.QWEN
        api_key = settings.qwen_api_key
        model_name = article.model_name or settings.qwen_default_model
    elif provider_str == "mistral":
        provider = LLMProviderType.MISTRAL
        api_key = settings.mistral_api_key
        model_name = article.model_name or settings.mistral_default_model
    elif provider_str == "groq":
        provider = LLMProviderType.GROQ
        api_key = getattr(settings, "groq_api_key", "") or ""
        model_name = article.model_name or getattr(settings, "groq_default_model", "llama-3.3-70b-versatile")
    else:
        provider = LLMProviderType.GEMINI
        api_key = settings.gemini_api_key
        model_name = article.model_name or settings.gemini_default_model

    if not api_key:
        raise ValueError(f"No API key configured for provider '{provider_str}'")

    llm_caller = LLMFactory.create_caller(
        provider=LLMProvider(provider.value),
        api_key=api_key,
        model=model_name,
        temperature=0.75,
        max_tokens=8192,
        timeout=settings.api_timeout,
    )
    llm = LangChainLLMWrapper(llm_caller=llm_caller)
    return llm, model_name, provider


def _chunk_text(chunk: Any) -> str:
    if isinstance(chunk, str):
        return chunk
    if hasattr(chunk, "text") and chunk.text:
        return str(chunk.text)
    if hasattr(chunk, "content") and chunk.content:
        return str(chunk.content)
    return str(chunk)


def stream_article_generation(
    *,
    article: ArticleProfile,
    customization: AssistantProfile,
    source_text: str,
    chapters: int,
    n_results: int,
    rag_system: Any,
) -> Iterator[str]:
    """Yield NDJSON lines: log, delta, chapter_start, chapter_done, done | error."""
    out_dir = os.path.join(settings.data_directory, "articles_output")
    os.makedirs(out_dir, exist_ok=True)

    file_id = str(uuid.uuid4())
    safe_name = "".join(c for c in article.name if c.isalnum() or c in (" ", "-", "_")).strip() or "article"
    safe_name = safe_name.replace(" ", "_")[:60]
    filename = f"{safe_name}_{file_id[:8]}.txt"
    filepath = os.path.join(out_dir, f"{file_id}.txt")

    try:
        llm, model_name, provider = _resolve_llm(article)
    except Exception as e:
        yield json.dumps({"type": "error", "message": str(e)}, ensure_ascii=False) + "\n"
        return

    yield json.dumps(
        {"type": "log", "message": f"Using model {model_name} ({provider.value})."},
        ensure_ascii=False,
    ) + "\n"

    combined_system = (
        f"## Base style (from customization: {customization.name})\n"
        f"{customization.system_prompt}\n\n"
        f"## Article instructions\n"
        f"{article.system_prompt}"
    )

    source_block = _truncate(source_text, SOURCE_MAX_CHARS)
    accumulated = ""
    full_parts: List[str] = []

    for c in range(1, chapters + 1):
        yield json.dumps(
            {"type": "log", "message": f"Starting chapter {c} of {chapters}..."},
            ensure_ascii=False,
        ) + "\n"

        rag_context = ""
        rag_used: Optional[str] = None
        if article.rag_collection:
            rag_used = article.rag_collection
            try:
                q = f"{article.system_prompt[:400]} Chapter {c} themes continuity"
                if source_block:
                    q += f"\n{source_block[:2000]}"
                results = rag_system.query_collection(
                    article.rag_collection,
                    q,
                    n_results,
                )
                if results:
                    rag_context = "\n\n".join(r["content"] for r in results[:n_results])
            except Exception as e:
                logger.warning("RAG query failed for chapter %s: %s", c, e)
                yield json.dumps(
                    {"type": "log", "message": f"RAG warning: {e}"},
                    ensure_ascii=False,
                ) + "\n"

        story_so_far = _truncate(accumulated, PREV_MAX_CHARS) if accumulated else ""

        user_parts = [
            f"You are writing a work in {chapters} chapters. This is chapter {c} of {chapters}.",
        ]
        if source_block:
            user_parts.append("### Optional source material\n" + source_block)
        if rag_context:
            user_parts.append(f"### Retrieved context (from '{rag_used}')\n{rag_context}")
        if story_so_far:
            user_parts.append("### Narrative so far (continuation)\n" + story_so_far)
        else:
            user_parts.append(
                "### Narrative so far\n(Opening chapter — establish tone, setting, and conflict as appropriate.)"
            )
        user_parts.append(
            f"Write **Chapter {c}** only. Use a clear chapter heading or title line. "
            "Do not summarize future chapters."
        )
        user_prompt = "\n\n".join(user_parts)

        full_prompt = f"{combined_system}\n\n---\n\n{user_prompt}"

        chapter_text = ""
        yield json.dumps({"type": "chapter_start", "chapter": c}, ensure_ascii=False) + "\n"
        try:
            stream_iter = getattr(llm, "stream", None)
            if stream_iter:
                for chunk in stream_iter(full_prompt):
                    piece = _chunk_text(chunk)
                    if piece:
                        chapter_text += piece
                        yield json.dumps(
                            {"type": "delta", "chapter": c, "text": piece},
                            ensure_ascii=False,
                        ) + "\n"
            if not chapter_text.strip():
                chapter_text = llm.invoke(full_prompt)
                if not isinstance(chapter_text, str):
                    chapter_text = str(chapter_text)
                yield json.dumps(
                    {"type": "log", "message": "Used non-streaming generation for this chapter."},
                    ensure_ascii=False,
                ) + "\n"
                yield json.dumps(
                    {"type": "delta", "chapter": c, "text": chapter_text},
                    ensure_ascii=False,
                ) + "\n"
        except Exception as e:
            logger.exception("Chapter %s generation failed", c)
            yield json.dumps({"type": "error", "message": str(e)}, ensure_ascii=False) + "\n"
            return

        accumulated += "\n\n" + chapter_text
        full_parts.append(f"## Chapter {c}\n\n{chapter_text.strip()}")
        yield json.dumps({"type": "chapter_done", "chapter": c}, ensure_ascii=False) + "\n"

    body = (
        f"# {article.name}\n\n"
        f"Generated with {model_name} ({provider.value}).\n\n"
        + "\n\n".join(full_parts)
    )
    with open(filepath, "w", encoding="utf-8") as f:
        f.write(body)

    yield json.dumps(
        {
            "type": "done",
            "file_id": file_id,
            "filename": filename,
            "model_used": model_name,
            "provider": provider.value,
        },
        ensure_ascii=False,
    ) + "\n"
