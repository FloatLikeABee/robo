"""
Articles module — saved profiles for multi-chapter text generation.
"""
import logging
import os
from typing import Dict, List, Optional

from tinydb import TinyDB, Query

from .config import settings
from .models import ArticleCreateRequest, ArticleProfile, ArticleUpdateRequest


class ArticleManager:
    """TinyDB-backed article profile storage."""

    def __init__(self) -> None:
        self.logger = logging.getLogger(__name__)
        self.articles: Dict[str, ArticleProfile] = {}

        os.makedirs(settings.data_directory, exist_ok=True)
        self.db_path = os.path.join(settings.data_directory, "articles.json")
        self.db = TinyDB(self.db_path)
        self.query = Query()

        self._load_articles()

    def _load_articles(self) -> None:
        try:
            for doc in self.db.all():
                try:
                    article_id = doc.get("id")
                    data = doc.get("profile", {})
                    if "id" in data:
                        del data["id"]
                    profile = ArticleProfile(id=article_id, **data)
                    self.articles[article_id] = profile
                except Exception as e:
                    self.logger.error("Failed to load article %s: %s", doc.get("id"), e)
            self.logger.info("Loaded %s article profiles", len(self.articles))
        except Exception as e:
            self.logger.error("Error loading articles: %s", e)

    def _save_articles(self) -> None:
        try:
            self.db.truncate()
            for aid, profile in self.articles.items():
                self.db.insert({"id": aid, "profile": profile.model_dump(exclude={"id"})})
            self.logger.info("Saved %s article profiles", len(self.articles))
        except Exception as e:
            self.logger.error("Error saving articles: %s", e)

    def _generate_id(self, name: str) -> str:
        base_id = name.strip().lower().replace(" ", "_")
        base_id = "".join(c for c in base_id if c.isalnum() or c in ("_", "-"))
        if not base_id:
            base_id = "article"
        candidate = base_id
        counter = 1
        while candidate in self.articles:
            candidate = f"{base_id}_{counter}"
            counter += 1
        return candidate

    def list_profiles(self) -> List[ArticleProfile]:
        return list(self.articles.values())

    def get_profile(self, article_id: str) -> Optional[ArticleProfile]:
        return self.articles.get(article_id)

    def create_profile(self, req: ArticleCreateRequest) -> str:
        article_id = self._generate_id(req.name)
        profile = ArticleProfile(
            id=article_id,
            name=req.name,
            description=req.description,
            customization_id=req.customization_id,
            system_prompt=req.system_prompt,
            rag_collection=req.rag_collection,
            llm_provider=req.llm_provider,
            model_name=req.model_name,
            default_chapters=req.default_chapters,
            metadata=req.metadata or {},
        )
        self.articles[article_id] = profile
        self._save_articles()
        self.logger.info("Created article profile: %s", article_id)
        return article_id

    def update_profile(self, article_id: str, req: ArticleUpdateRequest) -> bool:
        if article_id not in self.articles:
            return False
        self.articles[article_id] = ArticleProfile(
            id=article_id,
            name=req.name,
            description=req.description,
            customization_id=req.customization_id,
            system_prompt=req.system_prompt,
            rag_collection=req.rag_collection,
            llm_provider=req.llm_provider,
            model_name=req.model_name,
            default_chapters=req.default_chapters,
            metadata=req.metadata or {},
        )
        self._save_articles()
        return True

    def delete_profile(self, article_id: str) -> bool:
        if article_id not in self.articles:
            return False
        del self.articles[article_id]
        self._save_articles()
        return True
