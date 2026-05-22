# Data Model Draft

## Tables

- `users`: admin and normal user accounts.
- `api_keys`: user-created API keys for gateway access.
- `models`: public model catalog.
- `providers`: upstream provider channels.
- `provider_models`: provider-specific model routes.
- `request_logs`: per-request observability and usage.

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
