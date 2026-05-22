ALTER TABLE provider_models
    DROP COLUMN IF EXISTS pricing_json;

ALTER TABLE models
    DROP COLUMN IF EXISTS pricing_json;
