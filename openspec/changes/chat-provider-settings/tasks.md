# Tasks

## 1. Save and scan

- [x] 1.1 Add `user_ai_provider` and authenticated list, get, and save routes that never return the full key, and verify a test that GET masks the key and an empty update keeps it
- [x] 1.2 Add the scan route for OpenAI-compatible `/models` and Ollama `/api/tags`, and verify a test that a failed scan does not delete a saved row

## 2. Chat uses the saved provider

- [x] 2.1 Use the saved provider and model for that user's chat when both are set, and verify a test that a user with no row still uses the environment client

## 3. Session bar

- [x] 3.1 Add a settings button at the bottom of the session bar and a modal with provider, API URL, API key, Scan, and model choice, and verify the rail contains the control and Ollama can be saved without a key
- [x] 3.2 Run `openspec validate chat-provider-settings --strict` and verify the change is valid
