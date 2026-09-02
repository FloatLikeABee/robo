## Purpose

Removes the unused shared satellite chat package so operators use Morph AI as the only platform assistant surface.

## ADDED Requirements

### Requirement: platform-chat package is gone
The repository MUST NOT ship a `platform-chat/` npm package. Event Logs, Content Maker, Data Access, and Project MUST NOT show a PlatformChatDrawer-style assistant overlay. Morph AI chat MUST remain available as the system assistant.

#### Scenario: Event Logs has no shared chat drawer
- **WHEN** an operator opens Event Logs (FormsX) in MorphUtils
- **THEN** they do not see a floating platform-chat assistant drawer
- **AND** Morph AI remains reachable from the Morph AI app

#### Scenario: Package directory removed
- **WHEN** a developer inspects the repo root
- **THEN** there is no `platform-chat/` project folder
- **AND** app `package.json` files do not depend on `@robo/platform-chat`
