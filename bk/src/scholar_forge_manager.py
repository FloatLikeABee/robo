"""
ScholarForge — project CRUD for academic article & thesis composition.
"""
from __future__ import annotations

import logging
import os
import re
from datetime import datetime, timezone
from typing import Dict, List, Optional

from tinydb import Query, TinyDB

from .config import settings
from .models import (
    ScholarForgeCreateRequest,
    ScholarForgeDocumentMeta,
    ScholarForgeDocumentType,
    ScholarForgeProfile,
    ScholarForgeStatus,
    ScholarForgeUpdateRequest,
)


def _word_count(text: str) -> int:
    return len(re.findall(r"\S+", text or ""))


class ScholarForgeManager:
    """TinyDB-backed ScholarForge project storage."""

    def __init__(self) -> None:
        self.logger = logging.getLogger(__name__)
        self.projects: Dict[str, ScholarForgeProfile] = {}
        os.makedirs(settings.data_directory, exist_ok=True)
        self.db_path = os.path.join(settings.data_directory, "scholar_forge.json")
        self.db = TinyDB(self.db_path)
        self.query = Query()
        self._load_projects()

    def _load_projects(self) -> None:
        try:
            for doc in self.db.all():
                try:
                    pid = doc.get("id")
                    data = doc.get("profile", {})
                    if "id" in data:
                        del data["id"]
                    profile = ScholarForgeProfile(id=pid, **data)
                    self.projects[pid] = profile
                except Exception as e:
                    self.logger.error("Failed to load ScholarForge project %s: %s", doc.get("id"), e)
            self.logger.info("Loaded %s ScholarForge projects", len(self.projects))
        except Exception as e:
            self.logger.error("Error loading ScholarForge projects: %s", e)

    def _save_projects(self) -> None:
        try:
            self.db.truncate()
            for pid, profile in self.projects.items():
                self.db.insert({"id": pid, "profile": profile.model_dump(exclude={"id"})})
        except Exception as e:
            self.logger.error("Error saving ScholarForge projects: %s", e)

    def _generate_id(self, title: str) -> str:
        base = title.strip().lower().replace(" ", "_")
        base = "".join(c for c in base if c.isalnum() or c in ("_", "-")) or "scholar_forge"
        candidate = base[:48]
        counter = 1
        while candidate in self.projects:
            candidate = f"{base[:40]}_{counter}"
            counter += 1
        return candidate

    def _now(self) -> str:
        return datetime.now(timezone.utc).isoformat()

    def validate_detailed_prompt(self, text: str) -> None:
        min_words = settings.scholar_forge_min_prompt_words
        if _word_count(text) < min_words:
            raise ValueError(
                f"Detailed prompt must be at least {min_words} words "
                f"(currently {_word_count(text)})."
            )

    def validate_document_meta(
        self,
        document_type: ScholarForgeDocumentType,
        meta: ScholarForgeDocumentMeta,
    ) -> None:
        doc = document_type.value if hasattr(document_type, "value") else str(document_type)
        if doc in ("thesis", "dissertation"):
            missing = []
            if not (meta.author_name or "").strip():
                missing.append("author name")
            if not (meta.university or "").strip():
                missing.append("university / institution")
            if not (meta.department or "").strip():
                missing.append("department")
            if not (meta.degree_program or "").strip():
                missing.append("degree program")
            if not (meta.supervisor or "").strip():
                missing.append("supervisor / advisor")
            if missing:
                raise ValueError(
                    "Thesis and dissertation projects require: "
                    + ", ".join(missing)
                    + ". Fill in Author & document metadata."
                )
        elif doc == "article":
            if not (meta.author_name or "").strip():
                raise ValueError("Article projects require an author name.")

    def list_projects(self) -> List[ScholarForgeProfile]:
        return list(self.projects.values())

    def get_project(self, project_id: str) -> Optional[ScholarForgeProfile]:
        return self.projects.get(project_id)

    def save_project(self, profile: ScholarForgeProfile) -> None:
        profile.updated_at = self._now()
        self.projects[profile.id] = profile
        self._save_projects()

    def create_project(self, req: ScholarForgeCreateRequest) -> str:
        self.validate_detailed_prompt(req.detailed_prompt)
        self.validate_document_meta(req.document_type, req.document_meta)
        pid = self._generate_id(req.title)
        now = self._now()
        profile = ScholarForgeProfile(
            id=pid,
            subject=req.subject.strip(),
            title=req.title.strip(),
            short_intro=req.short_intro.strip(),
            detailed_prompt=req.detailed_prompt.strip(),
            recommended_sites=[s.strip() for s in req.recommended_sites if s and s.strip()],
            document_type=req.document_type,
            document_meta=req.document_meta,
            materials=req.materials,
            status=ScholarForgeStatus.DRAFT,
            created_at=now,
            updated_at=now,
        )
        self.projects[pid] = profile
        self._save_projects()
        return pid

    def update_project(self, project_id: str, req: ScholarForgeUpdateRequest) -> bool:
        if project_id not in self.projects:
            return False
        self.validate_detailed_prompt(req.detailed_prompt)
        self.validate_document_meta(req.document_type, req.document_meta)
        existing = self.projects[project_id]
        self.projects[project_id] = ScholarForgeProfile(
            id=project_id,
            subject=req.subject.strip(),
            title=req.title.strip(),
            short_intro=req.short_intro.strip(),
            detailed_prompt=req.detailed_prompt.strip(),
            recommended_sites=[s.strip() for s in req.recommended_sites if s and s.strip()],
            document_type=req.document_type,
            document_meta=req.document_meta,
            materials=req.materials,
            images=existing.images,
            rag_collection=existing.rag_collection,
            status=existing.status,
            clarification=existing.clarification,
            clarification_answers=existing.clarification_answers,
            structure=existing.structure,
            final_plan=existing.final_plan,
            section_cache=existing.section_cache,
            output_markdown_id=existing.output_markdown_id,
            output_pdf_id=existing.output_pdf_id,
            session_log=existing.session_log,
            metadata=existing.metadata,
            created_at=existing.created_at,
            updated_at=self._now(),
        )
        self._save_projects()
        return True

    def delete_project(self, project_id: str) -> bool:
        if project_id not in self.projects:
            return False
        del self.projects[project_id]
        self._save_projects()
        return True

    def append_log(self, project_id: str, message: str) -> None:
        profile = self.projects.get(project_id)
        if not profile:
            return
        profile.session_log = (profile.session_log or []) + [message]
        profile.updated_at = self._now()
        self._save_projects()
