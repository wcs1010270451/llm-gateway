ALTER TABLE provider_models
    DROP CONSTRAINT IF EXISTS provider_models_model_id_fkey;

DELETE FROM provider_models
WHERE model_id IS NULL;

ALTER TABLE provider_models
    ALTER COLUMN model_id SET NOT NULL;

ALTER TABLE provider_models
    ADD CONSTRAINT provider_models_model_id_fkey
        FOREIGN KEY (model_id)
        REFERENCES models(id)
        ON DELETE CASCADE;

COMMENT ON COLUMN provider_models.model_id IS '对应的对外模型 ID';
