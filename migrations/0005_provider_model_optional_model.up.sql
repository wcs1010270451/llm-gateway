ALTER TABLE provider_models
    DROP CONSTRAINT IF EXISTS provider_models_model_id_fkey;

ALTER TABLE provider_models
    ALTER COLUMN model_id DROP NOT NULL;

ALTER TABLE provider_models
    ADD CONSTRAINT provider_models_model_id_fkey
        FOREIGN KEY (model_id)
        REFERENCES models(id)
        ON DELETE SET NULL;

COMMENT ON COLUMN provider_models.model_id IS '关联的平台模型 ID；为空表示供应商已提供该上游模型，但平台暂未对外开放';
