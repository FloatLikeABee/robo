# Spec Delta

## ADDED Requirements

### Requirement: Twilight page luminance is dusk

The default twilight chat page background SHALL be visibly lighter than near-black (`#14101c` or darker). It SHALL remain dusk (not a near-white or beige daytime page). Body text SHALL remain light-on-dusk and meet WCAG 2.1 AA contrast against that page and against frosted assistant bubbles.

#### Scenario: Standalone chat after lift

- **WHEN** a user opens Dorph AI chat with the twilight preset
- **THEN** the page field reads as mid-dusk (lavender/slate), not night black
- **AND** primary message text remains readable on the page and on assistant bubbles

### Requirement: Stronger frosted glass on primary chrome

Header, session rail, composer, workspace pane, and assistant bubbles SHALL use frosted fills that let the dispersion mesh show through, with backdrop blur stronger than the first twilight pass. Glass borders SHALL include a light specular edge (brighter than the fill). User bubbles MAY stay more opaque for contrast.

#### Scenario: Chrome is glass, not a dark slab

- **WHEN** the chat header and composer are visible over the purple/orange mesh
- **THEN** mesh color is perceptible through those regions
- **AND** those regions still read as distinct bars (not fully clear)

#### Scenario: Reduced transparency still dusk

- **WHEN** `prefers-reduced-transparency: reduce` is active or backdrop-filter is unsupported
- **THEN** those regions use opaque dusk surfaces from the lifted token set
- **AND** they SHALL NOT fall back to near-black (`#14101c` or darker)
