# Backend

Backend service for authentication, model routing, upstream adapters, request logging, and usage statistics.

Read [ARCHITECTURE.md](ARCHITECTURE.md) before adding modules.

Planned modules:

- `cmd/server`: application entry point
- `internal/config`: configuration loading
- `internal/http`: route registration
- `internal/api`: HTTP handlers grouped by route domain
- `internal/auth`: admin/user/API-key auth
- `internal/repository`: database access split by domain
- `internal/routing`: public model to provider model resolution
- `internal/adapter`: upstream protocol adapters
- `internal/usage`: usage aggregation helpers

## Run

```powershell
..\scripts\dev-backend.ps1
```

Current endpoints:

- `GET /health`
- `GET /ready`
- `GET /api/admin/providers`
- `GET /api/admin/models`
- `GET /api/admin/models/:id/provider-models`
