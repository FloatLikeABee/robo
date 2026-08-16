# Remove tool system and merge adviser + customization into assistant

**Origin:** human

**Prompt:**
The current tool system (src/tools.py, ToolManager, ToolRegistry, individual tool wrappers) is not practical anymore. Remove it entirely from the codebase. Also merge the existing adviser and customization modules/concepts into a single new "assistant" module. The assistant module should unify personalized AI behaviors that were previously split between advisers and customizations.

**Tags:** refactor, architecture, assistant, tools-removal

<!-- OPENSPEC_IDEA_ENRICHMENT_START -->
## Enrichment Report

Generated: 2026-08-05T13:28:55Z

### Problem
TBD

### Proposed Direction
TBD

### Key Questions
- How many call sites depend on ToolManager/tool_manager across src/api.py, agent_manager, mcp_service, gathering_service, scholar_forge_service?
- Should the new Assistant module keep separate storage files for legacy adviser/customization data, or migrate into a single assistants.json?
- What replaces tool-enabled agents when ToolManager is removed — direct LLM calls only, or should agents still optionally bind to request_tools/db_tools?
- How does the frontend/API surface change: keep /advisers and /customizations endpoints or replace with /assistants?

### Feasibility
Feasibility: TBD

### T-Shirt Size
T-Shirt Size: TBD

### Size Justification
TBD

### Risks
- Removing tools.py may break agent_manager, mcp_service, gathering_service, and API endpoints that rely on ToolManager
- Merging adviser + customization data models requires careful migration to avoid data loss
- Frontend pages (AdviserManager.js, Customizations.js) depend on current endpoints and will need updates

### Suggested Next Step
Create an OpenSpec change with a baseline spec that defines the post-refactor architecture, then schedule implementation in smaller changes (remove tools, migrate adviser data, migrate customization data, unify API).
<!-- OPENSPEC_IDEA_ENRICHMENT_END -->
