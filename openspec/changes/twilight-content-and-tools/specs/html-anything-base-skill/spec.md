# Spec Delta

## MODIFIED Requirements

### Requirement: Compact skill uses hex colours only
The compact html-anything injection MUST name accent colours as hex values only. It MUST NOT place a colour word immediately before a hex value (for example `cyan #38bdf8`). The injected text MUST still stay at or under 512 tokens (Unicode runes divided by four, rounding up).

#### Scenario: Hex only under budget
- **WHEN** the compact injected skill text is inspected
- **THEN** it contains `#8b5cf6`, `#a78bfa`, and `#f97316`
- **AND** it does not contain `#38bdf8` or `#34d399`
- **AND** it does not contain `cyan #38bdf8` or `violet #a78bfa`
- **AND** the token estimate is 512 or fewer
