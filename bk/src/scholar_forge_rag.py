"""
ScholarForge — dedicated per-project RAG ingestion (isolated collections).
"""
from __future__ import annotations

import logging
from typing import Any, List, Optional

from .models import DataFormat, RAGDataInput, ScholarForgeMaterialInput, ScholarForgeProfile

logger = logging.getLogger(__name__)


def scholar_forge_collection_name(project_id: str) -> str:
    safe = project_id.lower().replace("-", "_")
    return f"scholar_forge_{safe}"[:63]


def ingest_materials_to_rag(
    rag_system: Any,
    project: ScholarForgeProfile,
    materials: List[ScholarForgeMaterialInput],
) -> str:
    """Create or reuse an isolated RAG collection and ingest all materials."""
    collection_name = scholar_forge_collection_name(project.id)
    for mat in materials:
        try:
            fmt = mat.format.lower()
            if fmt == "pdf_text":
                data_format = DataFormat.TXT
            else:
                data_format = DataFormat(fmt)
        except ValueError:
            data_format = DataFormat.TXT

        rag_input = RAGDataInput(
            name=mat.filename,
            description=mat.description or f"ScholarForge material for '{project.title}'",
            format=data_format,
            content=mat.content,
            tags=[project.id, "scholar_forge"],
            metadata={
                "project_id": project.id,
                "filename": mat.filename,
                "subject": project.subject,
            },
        )
        ok = rag_system.add_data_to_collection(collection_name, rag_input)
        if not ok:
            logger.warning(
                "Failed to ingest '%s' into collection '%s'",
                mat.filename,
                collection_name,
            )
    return collection_name


def query_project_rag(
    rag_system: Any,
    collection_name: str,
    query: str,
    n_results: int = 5,
) -> str:
    if not collection_name:
        return ""
    try:
        results = rag_system.query_collection(collection_name, query, n_results)
        if not results:
            return ""
        return "\n\n---\n\n".join(r["content"] for r in results[:n_results])
    except Exception as e:
        logger.warning("ScholarForge RAG query failed: %s", e)
        return ""
