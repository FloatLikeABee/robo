## Purpose

Project in MorphUtils must compile and load without the removed satellite chat package, using Morph AI as the only assistant.

## ADDED Requirements

### Requirement: Project UI has no platform-chat dependency
The Project frontend MUST compile and serve for MorphUtils embed without resolving `@robo/platform-chat` or mounting a PlatformChatDrawer-style overlay. The Project header MUST NOT include an **AI Assistant** control that opens that overlay. Morph AI MUST remain the system chat, reachable from the Morph AI app.

#### Scenario: Vite serves Project without the missing package
- **WHEN** an operator starts `morph-engi-ui` and opens MorphUtils Project
- **THEN** the Project UI loads without a failed import of `@robo/platform-chat/svelte` or `@robo/platform-chat/chat-drawer.css`

#### Scenario: No satellite drawer
- **WHEN** an operator uses Project inside MorphUtils
- **THEN** they do not see a floating Projects AI / platform-chat drawer
- **AND** they can still open Morph AI at its own origin for chat

### Requirement: Project embed is reachable on localhost
The Project Vite (or preview) server MUST listen on the documented MorphUtils Project embed port and MUST bind so `localhost` and `127.0.0.1` in an iframe can connect.

#### Scenario: Iframe can connect when the UI is running
- **WHEN** `morph-engi-ui` is listening on the Project embed URL
- **THEN** the MorphUtils Project iframe MUST NOT fail solely because the server bound IPv6-only or refused IPv4 `localhost`
