CREATE TABLE IF NOT EXISTS model_families (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(80) NOT NULL UNIQUE,
    display_name VARCHAR(120) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_model_families_status CHECK (status IN ('active', 'disabled'))
);

CREATE INDEX IF NOT EXISTS idx_model_families_status ON model_families(status);

DROP TRIGGER IF EXISTS trg_model_families_updated_at ON model_families;
CREATE TRIGGER trg_model_families_updated_at
BEFORE UPDATE ON model_families
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE model_families IS '模型系列字典，例如 gpt、claude、gemini';
COMMENT ON COLUMN model_families.id IS '模型系列主键 ID';
COMMENT ON COLUMN model_families.name IS '模型系列唯一标识，保存到 models.family';
COMMENT ON COLUMN model_families.display_name IS '模型系列显示名称';
COMMENT ON COLUMN model_families.status IS '模型系列状态：active 启用，disabled 禁用';
COMMENT ON COLUMN model_families.description IS '模型系列备注说明';
COMMENT ON COLUMN model_families.created_at IS '创建时间';
COMMENT ON COLUMN model_families.updated_at IS '更新时间';

INSERT INTO model_families (name, display_name, status)
SELECT DISTINCT family, family, 'active'
FROM models
WHERE family <> ''
ON CONFLICT (name) DO NOTHING;
