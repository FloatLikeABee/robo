## Purpose

Project in DorphUtils must compile and load without the removed satellite chat package, using Dorph AI as the only assistant.

## ADDED Requirements

### Requirement: Project UI has no platform-chat dependency
The Project frontend MUST compile and serve for DorphUtils embed without resolving `@robo/platform-chat` or mounting a PlatformChatDrawer-style overlay. The Project header MUST NOT include an **AI Assistant** control that opens that overlay. Dorph AI MUST remain the system chat, reachable from the Dorph AI app.

#### Scenario: Vite serves Project without the missing package
- **WHEN** an operator starts `dorph-engi-ui` and opens DorphUtils Project
- **THEN** the Project UI loads without a failed import of `@robo/platform-chat/svelte` or `@robo/platform-chat/chat-drawer.css`

#### Scenario: No satellite drawer
- **WHEN** an operator uses Project inside DorphUtils
- **THEN** they do not see a floating Projects AI / platform-chat drawer
- **AND** they can still open Dorph AI at its own origin for chat

### Requirement: Project embed is reachable on localhost
The Project Vite (or preview) server MUST listen on the documented DorphUtils Project embed port and MUST bind so `localhost` and `127.0.0.1` in an iframe can connect.

#### Scenario: Iframe can connect when the UI is running
- **WHEN** `dorph-engi-ui` is listening on the Project embed URL
- **THEN** the DorphUtils Project iframe MUST NOT fail solely because the server bound IPv6-only or refused IPv4 `localhost`
