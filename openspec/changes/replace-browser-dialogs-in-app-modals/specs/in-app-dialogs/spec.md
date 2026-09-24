## Purpose

Ensures confirms, warnings, and error notices in supported Morph stack products appear as themed in-app modals instead of browser-native alert/confirm dialogs.

## ADDED Requirements

### Requirement: No browser-native alert or confirm in product UI

Supported product frontends (Morph AI, MorphNotes, MorphUtils shell, Event Logs, Content Maker, Data Access, Project, AI tools) MUST NOT call `window.alert`, `window.confirm`, `window.prompt`, or unqualified global `alert` / `confirm` / `prompt` for user-facing messages. User feedback MUST render as an in-app modal or warning box owned by the product.

#### Scenario: Delete confirmation in Project

- **WHEN** the operator chooses to delete a project or project file
- **THEN** the app shows an in-app confirm modal with the action title and Cancel / Delete controls
- **AND** the browser does not show a native confirm dialog

#### Scenario: Error notice in AI tools

- **WHEN** loading agent details fails
- **THEN** the app shows an in-app notice modal (OK to dismiss)
- **AND** the browser does not show a native alert dialog

### Requirement: Confirm modals support destructive actions

In-app confirm modals MUST include a title, message body, a cancel control, and a confirm control. Destructive actions (delete, remove, irreversible) MUST use danger styling on the confirm control and explicit labels (e.g. “Delete”, not “OK”).

#### Scenario: Content Maker delete draft

- **WHEN** the operator confirms deleting a saved HTML draft
- **THEN** the modal names what will be deleted
- **AND** Cancel leaves the draft unchanged
- **AND** Delete proceeds only after explicit confirmation

### Requirement: Notice modals for non-destructive messages

Informational and error messages that previously used `alert` MUST use an in-app modal with a single dismiss control (e.g. “OK”). The modal MUST be keyboard-dismissible (Escape) and MUST NOT require a browser dialog.

#### Scenario: MorphNotes error in comments

- **WHEN** adding a comment fails
- **THEN** an in-app notice modal shows the error message
- **AND** the operator dismisses it without a native alert

### Requirement: Consistent dark theming

In-app dialogs MUST match the hosting product's dark UI (surface, border, text). They MUST NOT appear as unstyled browser chrome.

#### Scenario: Data Access delete row

- **WHEN** the operator sees a delete-row confirmation in Data Access
- **THEN** the modal uses the app's dark theme tokens
- **AND** it is visually consistent with other modals in that app

### Requirement: Event Logs retains existing confirm pattern

Event Logs MUST continue to use its existing in-app `ConfirmProvider` for confirmations. Any remaining native dialogs in that app MUST be migrated to the same pattern.

#### Scenario: Event Logs delete flow

- **WHEN** the operator confirms a destructive action in Event Logs
- **THEN** the existing in-app confirm overlay is used
- **AND** no native confirm appears
