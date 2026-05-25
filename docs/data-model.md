# Data Model Draft

## Tables

- `users`: admin and normal user accounts.
- `api_keys`: user-created API keys for gateway access.
- `models`: public model catalog.
- `providers`: upstream provider channels.
- `provider_models`: provider-specific model routes.
- `request_logs`: per-request observability and usage.

`api_keys.plain_key` keeps the historical database column name for compatibility, but stores an AES-GCM encrypted value. The user console decrypts only a credential owned by the authenticated user during an explicit reveal/copy request. Legacy plaintext values are upgraded to ciphertext when revealed.

## Manual Routing

`models.active_provider_model_id` points to the currently selected route in `provider_models`.

Example:

```text
model: sonnet-4.6
active_provider_model_id: 3

provider_model 1: Anthropic Official / claude-sonnet-4-6
provider_model 2: Claude Code Max / claude-sonnet-4-6
provider_model 3: Google Vertex AI / claude-sonnet-4-6
```
