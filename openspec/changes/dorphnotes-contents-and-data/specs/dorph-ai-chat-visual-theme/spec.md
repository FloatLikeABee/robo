# Spec Delta

## ADDED Requirements

### Requirement: Header product title is light on twilight glass

The Dorph AI header product title (“DORPH AI”) SHALL use the chat text token (light ink), not a clipped accent-color fill that matches the header bar. The title SHALL remain readable against the twilight glass header. The existing accent bar beside the title MAY keep purple/orange.

#### Scenario: Standalone chat header

- **WHEN** a user opens Dorph AI chat with the twilight preset
- **THEN** the words “DORPH AI” in the header are light (near `--chat-text`)
- **AND** they are visually distinct from the purple glass header background
