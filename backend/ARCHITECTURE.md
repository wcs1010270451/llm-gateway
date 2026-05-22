# Backend Architecture

The backend is organized by responsibility first, then by business domain.

```text
cmd/server/                 Process entry point. Builds config, DB, router, then starts HTTP.

internal/config/            Environment variable parsing and application config defaults.
internal/database/          GORM/PostgreSQL connection setup and pool tuning.
internal/entity/            Database entity structs. One file per domain table group.
internal/repository/        Database queries. One repository file per business domain.
internal/api/               HTTP handlers. One package per route group/domain.
internal/http/              Gin router composition only. No business logic here.
```

## Layer Rules

```text
cmd/server
  -> config
  -> database
  -> http router

http router
  -> api handlers
  -> repositories

api handlers
  -> repositories
  -> entities

repositories
  -> gorm
  -> entities
```

## Current Domains

```text
provider
  providers table
  upstream channel configuration

model
  models table
  provider_models table
  manual active provider model selection

system
  health and readiness checks
```

## Naming Rules

- `entity/<table>.go`: one database table struct per file.
- `repository/<domain>.go`: GORM query methods for that domain only.
- `api/<domain>/handler.go`: HTTP handler methods for that route group only.
- `http/router.go`: route registration and dependency wiring only.

If a file starts needing unrelated responsibilities, split it before adding more code.
