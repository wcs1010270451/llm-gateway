# Architecture

The gateway separates user-facing models from upstream provider models.

```text
Client request model
  -> public model
  -> active provider model
  -> provider channel adapter
  -> upstream API
```

## Core Concepts

- `models`: public models exposed to users, such as `sonnet-4.6`.
- `providers`: upstream channels, such as Anthropic Official, Claude Code Max, Google Vertex AI, or OpenAI Compatible Relay.
- `provider_models`: mapping from a public model to an upstream model on a provider.
- `models.active_provider_model_id`: first-version manual switch for choosing the current upstream path.
- `request_logs`: source of truth for provider/model usage statistics.

## MVP Routing

Version one uses manual routing only:

1. Resolve the requested public model.
2. Load its active provider model.
3. Dispatch to the adapter selected by `providers.adapter_type`.
4. Record request log and usage.

Automatic priority, failover, quota-aware, and cost-aware routing can be added later without changing the public API.
