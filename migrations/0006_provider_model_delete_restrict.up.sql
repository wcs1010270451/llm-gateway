ALTER TABLE provider_models
    DROP CONSTRAINT IF EXISTS provider_models_provider_id_fkey;

ALTER TABLE provider_models
    DROP CONSTRAINT IF EXISTS provider_models_model_id_fkey;

ALTER TABLE provider_models
    ADD CONSTRAINT provider_models_provider_id_fkey
        FOREIGN KEY (provider_id)
        REFERENCES providers(id)
        ON DELETE RESTRICT;

ALTER TABLE provider_models
    ADD CONSTRAINT provider_models_model_id_fkey
        FOREIGN KEY (model_id)
        REFERENCES models(id)
        ON DELETE RESTRICT;
