# llm-gateway

Personal and small-team LLM gateway for routing public model names to manually selected upstream provider models.

## Project Layout

```text
backend/      API server and adapter implementations
frontend/     Admin and user web console
migrations/   Database schema migrations
scripts/      Local development and deployment helper scripts
configs/      Example configuration files
docs/         Architecture and planning notes
deploy/       Deployment templates
tools/        Development utilities
```

## First MVP

- Admin and user accounts
- User API keys
- Providers and provider models
- Public models with a manually selected active provider model
- Request proxying
- Request logs and usage summaries

## Database

If you want this project to start its own local PostgreSQL:

```powershell
.\scripts\db-up.ps1
```

By default PostgreSQL is exposed on `127.0.0.1:55432` to avoid conflicts with an existing local PostgreSQL on `5432`.

If PostgreSQL is already running in another Docker container, skip `db-up.ps1` and point the migration scripts at that container:

```powershell
$env:POSTGRES_CONTAINER="your-postgres-container"
$env:POSTGRES_USER="your_user"
$env:POSTGRES_DB="your_database"
```

Apply the initial migration:

```powershell
.\scripts\db-migrate-up.ps1
```

List tables:

```powershell
.\scripts\db-tables.ps1
```
