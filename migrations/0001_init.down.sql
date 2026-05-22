DROP TABLE IF EXISTS request_logs;

ALTER TABLE models
    DROP CONSTRAINT IF EXISTS fk_models_active_provider_model;

DROP TABLE IF EXISTS provider_models;
DROP TABLE IF EXISTS models;
DROP TABLE IF EXISTS providers;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS set_updated_at();
