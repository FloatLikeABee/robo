"""Tests for unified Assistant module and tool system removal."""

import pytest


def test_tool_modules_removed():
    """Legacy tool modules should no longer exist."""
    import importlib
    with pytest.raises(ModuleNotFoundError):
        importlib.import_module("src.tools")
    with pytest.raises(ModuleNotFoundError):
        importlib.import_module("src.tool_registry")


def test_web_search_service_imports():
    from src.web_search_service import web_search, wikipedia_search
    assert callable(web_search)
    assert callable(wikipedia_search)


def test_assistant_manager_imports():
    from src.assistant_manager import AssistantManager
    from src.models import AssistantProfile, AssistantCreateRequest
    assert AssistantManager is not None
    assert AssistantProfile is not None
    assert AssistantCreateRequest is not None


def test_assistant_manager_crud():
    from src.assistant_manager import AssistantManager
    from src.agent_manager import AgentManager
    from src.rag_system import RAGSystem
    from src.models import AssistantCreateRequest, LLMProviderType

    rag = RAGSystem()
    agent_manager = AgentManager(rag)
    manager = AssistantManager(rag, agent_manager)

    # Clear any existing data for deterministic test
    for a in list(manager.list_assistants()):
        manager.delete_assistant(a.id)

    req = AssistantCreateRequest(
        name="test-assistant",
        system_prompt="You are a helpful test assistant.",
        llm_provider=LLMProviderType.GEMINI,
        model_name="gemini-2.5-flash",
    )
    assistant_id = manager.create_assistant(req)
    assert assistant_id is not None

    profile = manager.get_assistant(assistant_id)
    assert profile is not None
    assert profile.name == "test-assistant"

    updated = AssistantCreateRequest(
        name="test-assistant-renamed",
        system_prompt="You are still helpful.",
        llm_provider=LLMProviderType.GEMINI,
        model_name="gemini-2.5-flash",
    )
    assert manager.update_assistant(assistant_id, updated) is True
    assert manager.get_assistant(assistant_id).name == "test-assistant-renamed"

    assert manager.delete_assistant(assistant_id) is True
    assert manager.get_assistant(assistant_id) is None


def test_agent_manager_no_tool_dependency():
    from src.agent_manager import AgentManager
    from src.rag_system import RAGSystem
    rag = RAGSystem()
    # Should initialize without ToolManager
    manager = AgentManager(rag)
    assert manager is not None
