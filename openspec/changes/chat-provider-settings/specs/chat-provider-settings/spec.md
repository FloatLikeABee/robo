# Spec Delta

## Purpose

Lets a signed-in operator set the chat model from the session bar, including a cloud API or a local model, and scan that server for model names.

## ADDED Requirements

### Requirement: Settings sits at the bottom of the session bar
The chat sessions bar SHALL show a settings control at the bottom. Choosing it SHALL open a modal and SHALL NOT navigate away from the chat.

#### Scenario: Open settings
- **WHEN** an operator chooses the settings control on the session bar
- **THEN** a settings modal opens over the chat

### Requirement: The form accepts a provider, URL, and key
The modal SHALL list the existing providers, including the local Ollama provider. It SHALL accept an API URL and an API key. The local provider SHALL NOT require an API key. Saving SHALL store that choice for the signed-in user.

#### Scenario: Cloud provider
- **WHEN** the operator saves a cloud provider with an API URL and an API key
- **THEN** later chat requests from that user use that URL, key, and chosen model

#### Scenario: Local model
- **WHEN** the operator saves the local provider with an API URL and no key
- **THEN** the save succeeds
- **AND** chat uses that local server

#### Scenario: Nothing saved
- **WHEN** the user has not saved a provider
- **THEN** chat keeps using the server environment provider

### Requirement: Scan lists models from the given server
Scan SHALL call the given API URL with the given key and return the model ids that server reports. A failed scan SHALL show an error and SHALL NOT clear a previously saved choice.

#### Scenario: Scan succeeds
- **WHEN** the operator scans a reachable server
- **THEN** the modal lists that server's model ids
- **AND** the operator can choose one of them

#### Scenario: Scan fails
- **WHEN** the server cannot be reached or rejects the key
- **THEN** the modal shows the failure
- **AND** the saved provider is unchanged

### Requirement: The key is not shown again in full
A later read of the saved settings SHALL NOT return the full API key. It MAY show that a key is stored.

#### Scenario: Reopen settings
- **WHEN** the operator opens settings after a key was saved
- **THEN** the response does not include the full key
