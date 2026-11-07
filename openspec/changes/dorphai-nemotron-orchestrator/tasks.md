# Tasks

## 1. Provider and env (`pkg/dorphai`)

- [x] 1.1 Add `ProviderNVIDIA` to provider registry (base `https://integrate.api.nvidia.com/v1`, key `NVIDIA_API_KEY`, model default `nvidia/nemotron-3-ultra-550b-a55b`)
- [x] 1.2 Implement `LoadOrchestratorFromEnv` / `LoadCoordinatorFromEnv` with legacy fallback documented in README
- [x] 1.3 Send `chat_template_kwargs.enable_thinking` for NVIDIA compatible chat; add env `NVIDIA_ENABLE_THINKING` default true
- [x] 1.4 Add optional streaming API on Go client (reasoning vs content deltas) + httptest coverage
- [x] 1.5 Update `.env.example` with empty `NVIDIA_API_KEY`, orchestrator/coordinator provider vars (no secrets)

## 2. Memory storage (`dorph/db`, handlers)

- [x] 2.1 Add SQLite `agent_working_memory` table + migration
- [x] 2.2 Implement get/merge/patch with character cap; owner + session scoping
- [x] 2.3 Wire coordinator JSON `memory_patch` application in chat pipeline

## 3. Agent pipeline (`dorph/handlers`)

- [x] 3.1 Construct dual clients in `Handlers` startup from env
- [x] 3.2 Pre-turn coordinator call (skip simple chat); parse `coord_v1` JSON
- [x] 3.3 Build orchestrator system prompt: skills → lessons → working memory → plan
- [x] 3.4 Route `sub_agents` through existing worker prefetch + cap at 3
- [x] 3.5 Point `chatWithManagementTools` at orchestrator client only
- [x] 3.6 Point `distillSessionLesson` / `maybeHarvestSession` at coordinator when configured
- [x] 3.7 Integration tests with stub transports for both roles

## 4. API contract and UX (phased)

- [x] 4.1 Document dual-model env in `dorph/README.md` or `docs/agents/03-dorph.md`
- [x] 4.2 (Optional v1.1) SSE or chunked response on `POST /api/chat` for thinking + content
- [x] 4.3 (Follow-up) Align `pkg/dorphai-rs` / Project API with orchestrator-only profile

## 5. Verification

- [x] 5.1 `go test ./pkg/dorphai/...` and targeted `dorph` handler tests
- [ ] 5.2 Manual: set orchestrator+coordinator in `.env`, restart `dorph-api`, send tool + memory-heavy chat
- [x] 5.3 Confirm unset orchestrator env preserves legacy single-model behavior
