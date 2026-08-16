"""Unified Assistant manager: merges legacy Adviser and Customization behaviors."""

import json
import logging
import os
from datetime import datetime
from typing import Dict, List, Optional, Any

from tinydb import TinyDB, Query

from .config import settings
from .models import (
    AssistantCreateRequest,
    AssistantProfile,
    LLMProviderType,
    RAGDataInput,
    DataFormat,
    AgentConfig,
    AgentType,
)
from .rag_system import RAGSystem
from .llm_factory import LLMFactory, LLMProvider
from .agent_manager import AgentManager


class AssistantManager:
    """Manage Assistants: unified profiles combining adviser-like agents and customization-like prompts."""

    def __init__(
        self,
        rag_system: RAGSystem,
        agent_manager: AgentManager,
    ):
        self.logger = logging.getLogger(__name__)
        self.rag_system = rag_system
        self.agent_manager = agent_manager

        self.assistants: Dict[str, AssistantProfile] = {}

        os.makedirs(settings.data_directory, exist_ok=True)
        self.db_path = os.path.join(settings.data_directory, "assistants.json")
        self.db = TinyDB(self.db_path)
        self.query = Query()

        self._load_assistants()

    def _load_assistants(self) -> None:
        """Load assistant profiles from TinyDB."""
        try:
            docs = self.db.all()
            for doc in docs:
                try:
                    assistant_id = doc.get("id")
                    data = doc.get("profile", {})
                    if "id" in data:
                        del data["id"]
                    profile = AssistantProfile(id=assistant_id, **data)
                    self.assistants[assistant_id] = profile
                except Exception as e:
                    self.logger.error(f"Failed to load assistant {doc.get('id')}: {e}")
            self.logger.info(f"Loaded {len(self.assistants)} assistants")
        except Exception as e:
            self.logger.error(f"Error loading assistants: {e}")

    def _save_assistants(self) -> None:
        """Persist all assistant profiles to TinyDB."""
        try:
            self.db.truncate()
            for assistant_id, profile in self.assistants.items():
                self.db.insert(
                    {
                        "id": assistant_id,
                        "profile": profile.model_dump(exclude={"id"}),
                    }
                )
            self.logger.info(f"Saved {len(self.assistants)} assistants")
        except Exception as e:
            self.logger.error(f"Error saving assistants: {e}")

    def _generate_id(self, name: str) -> str:
        """Generate a stable, URL-friendly id from the assistant name."""
        base_id = name.strip().lower().replace(" ", "_")
        base_id = "".join(c for c in base_id if c.isalnum() or c in ("_", "-"))
        if not base_id:
            base_id = "assistant"

        candidate = base_id
        counter = 1
        while candidate in self.assistants:
            candidate = f"{base_id}_{counter}"
            counter += 1
        return candidate

    def list_assistants(self) -> List[AssistantProfile]:
        return list(self.assistants.values())

    def get_assistant(self, assistant_id: str) -> Optional[AssistantProfile]:
        return self.assistants.get(assistant_id)

    def delete_assistant(self, assistant_id: str) -> bool:
        """Delete an assistant and its underlying agent (RAG collections are preserved)."""
        if assistant_id not in self.assistants:
            return False
        try:
            profile = self.assistants[assistant_id]
            agent_id = profile.agent_id
            if agent_id:
                try:
                    self.agent_manager.delete_agent(agent_id)
                except Exception as e:
                    self.logger.warning(
                        f"Failed to delete underlying agent '{agent_id}' for assistant '{assistant_id}': {e}"
                    )
            del self.assistants[assistant_id]
            self.db.remove(self.query.id == assistant_id)
            self.logger.info(f"Deleted assistant: {assistant_id}")
            return True
        except Exception as e:
            self.logger.error(f"Error deleting assistant {assistant_id}: {e}")
            return False

    def _default_provider_and_model(
        self, provider_override: Optional[LLMProviderType], model_override: Optional[str]
    ):
        """Resolve provider enum, API key, and model name with sensible defaults."""
        provider_str = (
            provider_override.value if provider_override else settings.default_llm_provider
        )

        if provider_str == "gemini":
            provider = LLMProviderType.GEMINI
            api_key = settings.gemini_api_key
            model_name = model_override or settings.gemini_default_model
        elif provider_str == "qwen":
            provider = LLMProviderType.QWEN
            api_key = settings.qwen_api_key
            model_name = model_override or settings.qwen_default_model
        elif provider_str == "mistral":
            provider = LLMProviderType.MISTRAL
            api_key = settings.mistral_api_key
            model_name = model_override or settings.mistral_default_model
        elif provider_str == "groq":
            provider = LLMProviderType.GROQ
            api_key = getattr(settings, "groq_api_key", "")
            model_name = model_override or getattr(
                settings, "groq_default_model", "llama-3.3-70b-versatile"
            )
        else:
            provider = LLMProviderType.GEMINI
            api_key = settings.gemini_api_key
            model_name = model_override or settings.gemini_default_model

        return provider, api_key, model_name

    def _normalize_prompt_and_description(
        self,
        draft_prompt: str,
        draft_description: Optional[str],
        provider_override: Optional[LLMProviderType],
        model_override: Optional[str],
    ) -> Dict[str, str]:
        """Use LLM to clean up system prompt and ensure we have a solid description."""
        provider, api_key, model_name = self._default_provider_and_model(
            provider_override, model_override
        )

        try:
            llm_caller = LLMFactory.create_caller(
                provider=LLMProvider(provider.value),
                api_key=api_key,
                model=model_name,
                temperature=0.3,
                max_tokens=1024,
            )
        except Exception as e:
            self.logger.warning(
                f"Failed to initialize LLM for assistant normalization, using raw prompt/description. Error: {e}"
            )
            return {
                "system_prompt": draft_prompt,
                "description": draft_description
                or self._fallback_description_from_prompt(draft_prompt),
            }

        payload = {
            "prompt": draft_prompt,
            "description": draft_description or "",
        }
        normalization_instructions = (
            "You are an AI assistant that rewrites agent system prompts and descriptions "
            "to be clear, professional, safe, and aligned with best practices.\n\n"
            "Given the following JSON with a draft system prompt and an optional description, "
            "you MUST respond with a single, valid JSON object ONLY, with this exact shape:\n"
            '{\n  "system_prompt": "cleaned and improved system prompt text",\n'
            '  "description": "short, user-facing description of what this assistant does"\n}\n\n'
            "- Do not add explanations.\n"
            "- Do not add extra fields.\n"
            "- Keep the description concise (1–2 sentences).\n"
        )

        full_prompt = (
            f"{normalization_instructions}\n\n"
            f"INPUT_JSON:\n{json.dumps(payload, ensure_ascii=False)}"
        )

        try:
            raw = llm_caller.generate(full_prompt)
            cleaned = raw.strip()
            if cleaned.startswith("```"):
                cleaned = cleaned.strip("`")
                if "\n" in cleaned:
                    cleaned = cleaned.split("\n", 1)[1]
            data = json.loads(cleaned)
            system_prompt = (
                str(data.get("system_prompt")).strip() if data.get("system_prompt") else draft_prompt
            )
            description = (
                str(data.get("description")).strip()
                if data.get("description")
                else draft_description or self._fallback_description_from_prompt(draft_prompt)
            )
            return {"system_prompt": system_prompt, "description": description}
        except Exception as e:
            self.logger.warning(
                f"Failed to parse normalization response, using raw prompt/description. Error: {e}"
            )
            return {
                "system_prompt": draft_prompt,
                "description": draft_description
                or self._fallback_description_from_prompt(draft_prompt),
            }

    def _fallback_description_from_prompt(self, prompt: str) -> str:
        snippet = prompt.strip().replace("\n", " ")
        if len(snippet) > 180:
            snippet = snippet[:177].rstrip() + "..."
        return f"Assistant configured with system instructions: {snippet}"

    def _ingest_files_to_collection(
        self, assistant_id: str, assistant_name: str, files: List[Any]
    ) -> Optional[str]:
        """Create (or reuse) an assistant-specific RAG collection and ingest uploaded files."""
        if not files:
            return None

        collection_name = f"assistant_{assistant_id}_kb"
        for file in files:
            try:
                # Handle both AdviserFileInput and plain dict shapes
                filename = getattr(file, "filename", file.get("filename", "unknown"))
                content = getattr(file, "content", file.get("content", ""))
                description = getattr(file, "description", file.get("description"))
                fmt = getattr(file, "format", file.get("format", "txt"))
                if isinstance(fmt, str):
                    fmt = DataFormat(fmt)

                rag_input = RAGDataInput(
                    name=filename,
                    description=description
                    or f"Base knowledge file for assistant '{assistant_name}'",
                    format=fmt,
                    content=content,
                    tags=[assistant_id, "assistant"],
                    metadata={"assistant_id": assistant_id, "filename": filename},
                )
                success = self.rag_system.add_data_to_collection(collection_name, rag_input)
                if not success:
                    self.logger.warning(
                        f"Failed to add data from file '{filename}' to collection '{collection_name}'"
                    )
            except Exception as e:
                self.logger.error(
                    f"Error ingesting file for assistant '{assistant_id}': {e}"
                )

        return collection_name

    def _build_agent_config(
        self,
        name: str,
        description: str,
        system_prompt: str,
        provider: LLMProviderType,
        model_name: str,
        rag_collections: List[str],
    ) -> AgentConfig:
        """Create an AgentConfig for this assistant using RAG and direct LLM (tools removed)."""
        return AgentConfig(
            name=name,
            description=description,
            agent_type=AgentType.HYBRID,
            llm_provider=provider,
            model_name=model_name,
            temperature=0.7,
            max_tokens=8192,
            rag_collections=rag_collections,
            tools=[],
            system_prompt=system_prompt,
            system_prompt_data=None,
            is_active=True,
        )

    def create_assistant(self, req: AssistantCreateRequest) -> str:
        """Create a new assistant, including base RAG data and underlying agent."""
        if not req.system_prompt or not req.system_prompt.strip():
            raise ValueError("system_prompt is required to create an assistant")

        assistant_id = self._generate_id(req.name)

        normalized = self._normalize_prompt_and_description(
            draft_prompt=req.system_prompt,
            draft_description=req.description,
            provider_override=req.llm_provider,
            model_override=req.model_name,
        )
        system_prompt = normalized["system_prompt"]
        final_description = normalized["description"]

        provider, _, resolved_model_name = self._default_provider_and_model(
            req.llm_provider, req.model_name
        )

        base_collection = self._ingest_files_to_collection(
            assistant_id=assistant_id,
            assistant_name=req.name,
            files=req.files,
        )

        rag_collections: List[str] = list(req.existing_rag_collections or [])
        if base_collection:
            rag_collections.append(base_collection)
        if req.rag_collection and req.rag_collection not in rag_collections:
            rag_collections.append(req.rag_collection)
        seen: set = set()
        deduped_rag_collections: List[str] = []
        for col in rag_collections:
            if col and col not in seen:
                seen.add(col)
                deduped_rag_collections.append(col)

        agent_config = self._build_agent_config(
            name=req.name,
            description=final_description,
            system_prompt=system_prompt,
            provider=provider,
            model_name=resolved_model_name,
            rag_collections=deduped_rag_collections,
        )
        agent_id = self.agent_manager.create_agent(agent_config)

        profile = AssistantProfile(
            id=assistant_id,
            name=req.name,
            description=final_description,
            system_prompt=system_prompt,
            rag_collections=deduped_rag_collections,
            base_collection=base_collection,
            llm_provider=provider,
            model_name=resolved_model_name,
            agent_id=agent_id,
            request_tool_id=req.request_tool_id,
            db_tool_id=req.db_tool_id,
            tool_response_mode=req.tool_response_mode,
            metadata=req.metadata,
        )

        self.assistants[assistant_id] = profile
        self._save_assistants()
        self.logger.info(
            f"Created assistant '{assistant_id}' with agent '{agent_id}' and {len(deduped_rag_collections)} RAG collections"
        )
        return assistant_id

    def update_assistant(self, assistant_id: str, req: AssistantCreateRequest) -> bool:
        """Update an existing assistant. New files are appended to the same base collection."""
        if assistant_id not in self.assistants:
            return False

        existing = self.assistants[assistant_id]

        normalized = self._normalize_prompt_and_description(
            draft_prompt=req.system_prompt,
            draft_description=req.description or existing.description,
            provider_override=req.llm_provider or existing.llm_provider,
            model_override=req.model_name or existing.model_name,
        )
        system_prompt = normalized["system_prompt"]
        final_description = normalized["description"]

        provider, _, resolved_model_name = self._default_provider_and_model(
            req.llm_provider or existing.llm_provider, req.model_name or existing.model_name
        )

        base_collection = existing.base_collection or f"assistant_{assistant_id}_kb"
        if req.files:
            for file in req.files:
                try:
                    filename = getattr(file, "filename", file.get("filename", "unknown"))
                    content = getattr(file, "content", file.get("content", ""))
                    description = getattr(file, "description", file.get("description"))
                    fmt = getattr(file, "format", file.get("format", "txt"))
                    if isinstance(fmt, str):
                        fmt = DataFormat(fmt)

                    rag_input = RAGDataInput(
                        name=filename,
                        description=description
                        or f"Base knowledge file for assistant '{req.name}'",
                        format=fmt,
                        content=content,
                        tags=[assistant_id, "assistant"],
                        metadata={"assistant_id": assistant_id, "filename": filename},
                    )
                    self.rag_system.add_data_to_collection(base_collection, rag_input)
                except Exception as e:
                    self.logger.error(
                        f"Error ingesting file for assistant '{assistant_id}' on update: {e}"
                    )

        fields_set = getattr(req, "model_fields_set", None) or getattr(req, "__fields_set__", set())
        if "existing_rag_collections" in fields_set:
            rag_collections: List[str] = list(req.existing_rag_collections or [])
        else:
            rag_collections = list(existing.rag_collections or [])
        if base_collection:
            rag_collections.append(base_collection)
        if req.rag_collection and req.rag_collection not in rag_collections:
            rag_collections.append(req.rag_collection)
        seen: set = set()
        deduped_rag_collections: List[str] = []
        for col in rag_collections:
            if col and col not in seen:
                seen.add(col)
                deduped_rag_collections.append(col)

        agent_config = self._build_agent_config(
            name=req.name,
            description=final_description,
            system_prompt=system_prompt,
            provider=provider,
            model_name=resolved_model_name,
            rag_collections=deduped_rag_collections,
        )

        agent_id = existing.agent_id
        if agent_id:
            try:
                updated = self.agent_manager.update_agent(agent_id, agent_config)
                if not updated:
                    agent_id = self.agent_manager.create_agent(agent_config)
            except Exception as e:
                self.logger.warning(
                    f"Agent update failed for assistant '{assistant_id}', recreating agent. Error: {e}"
                )
                agent_id = self.agent_manager.create_agent(agent_config)
        else:
            agent_id = self.agent_manager.create_agent(agent_config)

        profile = AssistantProfile(
            id=assistant_id,
            name=req.name,
            description=final_description,
            system_prompt=system_prompt,
            rag_collections=deduped_rag_collections,
            base_collection=base_collection,
            llm_provider=provider,
            model_name=resolved_model_name,
            agent_id=agent_id,
            request_tool_id=req.request_tool_id or existing.request_tool_id,
            db_tool_id=req.db_tool_id or existing.db_tool_id,
            tool_response_mode=req.tool_response_mode or existing.tool_response_mode,
            metadata=req.metadata or existing.metadata,
        )

        self.assistants[assistant_id] = profile
        self._save_assistants()
        self.logger.info(
            f"Updated assistant '{assistant_id}' with agent '{agent_id}' and {len(deduped_rag_collections)} RAG collections"
        )
        return True

    def migrate_from_legacy(
        self,
        adviser_manager=None,
        customization_manager=None,
    ) -> Dict[str, int]:
        """One-time migration from legacy adviser and customization managers."""
        migrated = {"advisers": 0, "customizations": 0}

        if adviser_manager:
            for adviser in adviser_manager.list_advisers():
                try:
                    profile = AssistantProfile(
                        id=f"adviser_{adviser.id}",
                        name=adviser.name,
                        description=adviser.description,
                        system_prompt=adviser.system_prompt,
                        rag_collections=adviser.rag_collections,
                        base_collection=adviser.base_collection,
                        llm_provider=adviser.llm_provider,
                        model_name=adviser.model_name,
                        agent_id=adviser.agent_id,
                        request_tool_id=None,
                        db_tool_id=None,
                        tool_response_mode="raw",
                        metadata={"source": "adviser", "original_id": adviser.id},
                    )
                    self.assistants[profile.id] = profile
                    migrated["advisers"] += 1
                except Exception as e:
                    self.logger.error(f"Failed to migrate adviser {adviser.id}: {e}")

        if customization_manager:
            for customization in customization_manager.list_profiles():
                try:
                    rag_collections = []
                    if customization.rag_collection:
                        rag_collections.append(customization.rag_collection)
                    profile = AssistantProfile(
                        id=f"customization_{customization.id}",
                        name=customization.name,
                        description=customization.description,
                        system_prompt=customization.system_prompt,
                        rag_collections=rag_collections,
                        base_collection=None,
                        llm_provider=customization.llm_provider,
                        model_name=customization.model_name,
                        agent_id=None,
                        request_tool_id=customization.request_tool_id,
                        db_tool_id=customization.db_tool_id,
                        tool_response_mode=customization.tool_response_mode,
                        metadata={"source": "customization", "original_id": customization.id},
                    )
                    self.assistants[profile.id] = profile
                    migrated["customizations"] += 1
                except Exception as e:
                    self.logger.error(f"Failed to migrate customization {customization.id}: {e}")

        if migrated["advisers"] or migrated["customizations"]:
            self._save_assistants()
            self.logger.info(f"Migrated {migrated}")
        return migrated

    def _ensure_underlying_agent(self, profile: AssistantProfile) -> str:
        """Return a runnable agent id for this assistant, recreating the agent if missing."""
        agent_id = (profile.agent_id or profile.id or "").strip()
        if agent_id and self.agent_manager.get_agent(agent_id):
            return agent_id

        provider, _, model_name = self._default_provider_and_model(
            profile.llm_provider, profile.model_name
        )
        agent_config = self._build_agent_config(
            name=profile.name,
            description=profile.description or "",
            system_prompt=profile.system_prompt or "",
            provider=provider,
            model_name=model_name,
            rag_collections=list(profile.rag_collections or []),
        )
        desired_id = agent_id or profile.id
        created_id = self.agent_manager.create_agent(
            agent_config, fixed_agent_id=desired_id
        )
        if profile.agent_id != created_id:
            profile.agent_id = created_id
            self.assistants[profile.id] = profile
            self._save_assistants()
            self.logger.info(
                f"Recreated missing agent '{created_id}' for assistant '{profile.id}'"
            )
        return created_id

    async def run_assistant(
        self, assistant_id: str, query: str, context: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Run the underlying agent for an assistant."""
        profile = self.get_assistant(assistant_id)
        if not profile:
            raise ValueError(f"Assistant {assistant_id} not found")

        agent_id = self._ensure_underlying_agent(profile)
        result = await self.agent_manager.run_agent(agent_id, query, context)
        result["assistant_id"] = assistant_id
        return result

