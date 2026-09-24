## Why

Twenty research rounds take too long and the “final” document is still a round-by-round merge. Operators need a shorter run and a single conclusion that reads like a finished essay or thesis: the essence of every round absorbed into one argument, not a stapled packet of Round 1…Round N.

## What Changes

- New MorphNotes Research jobs run **five** online verified rounds instead of twenty.
- After the last round, the conclusion is a **thesis-style synthesis**: one professional document that reorganizes and absorbs the best-supported material from all rounds. Round order MUST NOT be the spine of the conclusion.
- A critic polish pass follows the first synthesis draft so the stored Markdown/HTML is a finished piece, not a light copy-edit of concatenated rounds.
- Per-round pieces stay visible on the Rounds tab (working notes). The Markdown/HTML tabs show only the synthesized conclusion.
- Progress copy and `round_count` follow five for new jobs. Jobs already created keep the round target they started with.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `morphnotes-research`: Round count is five, not twenty. The stored conclusion is a synthesized thesis-quality document, not pieces concatenated (or refined) in research-step order.

## Impact

- MorphNotes API worker: `morph/handlers/tran_research.go` (round loop, refine prompts, fallback assemble), tests in `tran_research_test.go`.
- Optional SQLite column so existing jobs keep their original round target (`round_target` or equivalent).
- MorphNotes Research UI: `Research.js` copy (`N/5`, subtitle), progress bar; Markdown tab still `MarkdownEditor`.
- No MorphUtils, ComposerX, or theme-switch work. Same in-process worker, `webresearch.Gather`, and Morph-native HTML publish.
