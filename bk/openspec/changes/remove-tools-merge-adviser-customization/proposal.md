# Remove tool system and merge adviser + customization into assistant

## Motivation

The legacy tool system (`src/tools.py`, `ToolManager`, `ToolRegistry`, and built-in tool wrappers) has become impractical: it is tightly coupled to agents, MCP, gathering, scholar-forge, and the API, yet most tool capabilities duplicate LLM reasoning or external APIs we can call directly. Removing it simplifies the architecture.

At the same time, the current `adviser` and `customization` concepts overlap heavily (both define system prompts + optional RAG context). Merging them into a single `assistant` module gives users one place to configure personalized AI behaviors.

## Scope

This change is **planning-only**. Implementation will be split into smaller follow-up changes to keep the refactor reviewable and safe.

## Plan

1. **Spec the target state** — baseline `assistant-module` spec defines the unified Assistant module and tool-less architecture.
2. **Phase 1: Remove tool system**
   - Delete `src/tools.py` and `src/tool_registry.py`
   - Remove `ToolManager` dependency from `agent_manager`, `mcp_service`, `gathering_service`, `scholar_forge_service`, `api.py`
   - Convert tool-enabled agents to direct LLM calls
   - Update MCP `tools/list` and `tools/call` to return/call only external host tools
3. **Phase 2: Introduce Assistant module**
   - Create `src/assistant_manager.py` and `AssistantProfile` model
   - Add `src/assistants.py` service facade
   - Add `GET|POST|PUT|DELETE /assistants` endpoints
4. **Phase 3: Migrate data**
   - One-time migration from `advisers.json` + `customizations.json` → `assistants.json`
   - Preserve source metadata (`source: adviser` / `source: customization`)
5. **Phase 4: Remove legacy modules and endpoints**
   - Delete `src/adviser_manager.py`, `src/customization.py`, `src/customization_tools.py`
   - Remove `/advisers` and `/customizations` endpoints
   - Replace frontend pages with an `Assistants` page

## Acceptance Criteria

- `ToolManager` no longer exists anywhere in the codebase.
- `src/tools.py` and `src/tool_registry.py` are deleted.
- `AssistantProfile` is the single profile type for personalized behaviors.
- `/assistants` endpoints serve all migrated adviser and customization data.
- All existing tests pass; new tests cover Assistant CRUD and migration.
