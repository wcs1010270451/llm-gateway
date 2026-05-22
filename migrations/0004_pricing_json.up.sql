ALTER TABLE models
    ADD COLUMN IF NOT EXISTS pricing_json JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE provider_models
    ADD COLUMN IF NOT EXISTS pricing_json JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN models.pricing_json IS '平台模型计价规则 JSON，用于用户侧用量统计和结算';
COMMENT ON COLUMN provider_models.pricing_json IS '供应商模型成本规则 JSON，用于上游成本统计';
