"""Video Story Generator — project/scene CRUD with TinyDB persistence."""
from __future__ import annotations

import logging
import os
import uuid
from datetime import datetime
from typing import Dict, List, Optional

from tinydb import Query, TinyDB

from .config import settings
from .models import (
    VideoStoryCreateRequest,
    VideoStoryProfile,
    VideoStoryScene,
    VideoStorySceneStatus,
    VideoStoryStatus,
    VideoStoryUpdateRequest,
)


class VideoStoryManager:
    def __init__(self) -> None:
        self.logger = logging.getLogger(__name__)
        self.projects: Dict[str, VideoStoryProfile] = {}
        os.makedirs(settings.data_directory, exist_ok=True)
        self.db_path = os.path.join(settings.data_directory, "video_stories.json")
        self.db = TinyDB(self.db_path)
        self.query = Query()
        self._load_projects()

    def _load_projects(self) -> None:
        try:
            for doc in self.db.all():
                pid = doc.get("id")
                data = doc.get("profile", {})
                if "id" in data:
                    del data["id"]
                self.projects[pid] = VideoStoryProfile(id=pid, **data)
            self.logger.info("Loaded %s video story projects", len(self.projects))
        except Exception as e:
            self.logger.error("Error loading video stories: %s", e)

    def _save_projects(self) -> None:
        try:
            self.db.truncate()
            for pid, profile in self.projects.items():
                self.db.insert({"id": pid, "profile": profile.model_dump(exclude={"id"})})
        except Exception as e:
            self.logger.error("Error saving video stories: %s", e)

    def _generate_id(self, title: str) -> str:
        base = title.strip().lower().replace(" ", "_")
        base = "".join(c for c in base if c.isalnum() or c in ("_", "-")) or "story"
        candidate = base
        n = 1
        while candidate in self.projects:
            candidate = f"{base}_{n}"
            n += 1
        return candidate

    def project_dir(self, project_id: str) -> str:
        path = os.path.join(settings.data_directory, settings.video_story_output_dir, project_id)
        os.makedirs(path, exist_ok=True)
        os.makedirs(os.path.join(path, "images"), exist_ok=True)
        os.makedirs(os.path.join(path, "videos"), exist_ok=True)
        os.makedirs(os.path.join(path, "shared"), exist_ok=True)
        return path

    def list_projects(self) -> List[VideoStoryProfile]:
        return list(self.projects.values())

    def get_project(self, project_id: str) -> Optional[VideoStoryProfile]:
        return self.projects.get(project_id)

    def create_project(self, req: VideoStoryCreateRequest) -> str:
        pid = self._generate_id(req.title)
        now = datetime.utcnow().isoformat() + "Z"
        scenes: List[VideoStoryScene] = []
        for i, raw in enumerate(req.scenes or []):
            scenes.append(
                VideoStoryScene(
                    id=str(uuid.uuid4()),
                    order=i + 1,
                    title=str(raw.get("title") or f"Scene {i + 1}"),
                    user_prompt=str(raw.get("user_prompt") or raw.get("prompt") or ""),
                )
            )
        profile = VideoStoryProfile(
            id=pid,
            title=req.title.strip(),
            description=req.description.strip(),
            story_context=req.story_context.strip(),
            scenes=scenes,
            status=VideoStoryStatus.DRAFT,
            created_at=now,
            updated_at=now,
        )
        self.projects[pid] = profile
        self.project_dir(pid)
        self._save_projects()
        return pid

    def update_project(self, project_id: str, req: VideoStoryUpdateRequest) -> bool:
        profile = self.projects.get(project_id)
        if not profile:
            return False
        profile.title = req.title.strip()
        profile.description = req.description.strip()
        profile.story_context = req.story_context.strip()
        profile.scenes = req.scenes
        if req.cast is not None:
            profile.cast = req.cast
        if req.world is not None:
            profile.world = req.world
        if req.style_bible is not None:
            profile.style_bible = req.style_bible.strip()
        profile.updated_at = datetime.utcnow().isoformat() + "Z"
        self._save_projects()
        return True

    def save_project(self, profile: VideoStoryProfile) -> None:
        profile.updated_at = datetime.utcnow().isoformat() + "Z"
        self.projects[profile.id] = profile
        self._save_projects()

    def delete_project(self, project_id: str) -> bool:
        if project_id not in self.projects:
            return False
        del self.projects[project_id]
        self._save_projects()
        return True
