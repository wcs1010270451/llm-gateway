CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name VARCHAR(120) NOT NULL DEFAULT '',
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_users_role CHECK (role IN ('admin', 'user')),
    CONSTRAINT chk_users_status CHECK (status IN ('active', 'disabled'))
);

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE users IS '系统用户表，包含管理员和普通用户';
COMMENT ON COLUMN users.id IS '用户主键 ID';
COMMENT ON COLUMN users.email IS '用户邮箱，登录账号';
COMMENT ON COLUMN users.password_hash IS '密码哈希';
COMMENT ON COLUMN users.display_name IS '显示名称';
COMMENT ON COLUMN users.role IS '用户角色：admin 表示管理员，user 表示普通用户';
COMMENT ON COLUMN users.status IS '用户状态：active 启用，disabled 禁用';
COMMENT ON COLUMN users.last_login_at IS '最近登录时间';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '更新时间';

CREATE TABLE api_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    plain_key TEXT NOT NULL DEFAULT '',
    masked_key VARCHAR(80) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    rpm_limit INTEGER NOT NULL DEFAULT 0,
    daily_request_limit INTEGER NOT NULL DEFAULT 0,
    daily_token_limit INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NULL,
    last_used_at TIMESTAMPTZ NULL,
    last_error_at TIMESTAMPTZ NULL,
    last_error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_api_keys_status CHECK (status IN ('active', 'disabled')),
    CONSTRAINT chk_api_keys_rpm_limit CHECK (rpm_limit >= 0),
    CONSTRAINT chk_api_keys_daily_request_limit CHECK (daily_request_limit >= 0),
    CONSTRAINT chk_api_keys_daily_token_limit CHECK (daily_token_limit >= 0)
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_status ON api_keys(status);
CREATE INDEX idx_api_keys_expires_at ON api_keys(expires_at);

CREATE TRIGGER trg_api_keys_updated_at
BEFORE UPDATE ON api_keys
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE api_keys IS '用户调用网关时使用的 API Key';
COMMENT ON COLUMN api_keys.id IS 'API Key 主键 ID';
COMMENT ON COLUMN api_keys.user_id IS '所属用户 ID';
COMMENT ON COLUMN api_keys.name IS 'API Key 显示名称';
COMMENT ON COLUMN api_keys.key_hash IS 'API Key 哈希值，仅用于鉴权匹配';
COMMENT ON COLUMN api_keys.masked_key IS '脱敏后的 API Key 展示值';
COMMENT ON COLUMN api_keys.status IS 'Key 状态：active 启用，disabled 禁用';
COMMENT ON COLUMN api_keys.rpm_limit IS '每分钟请求数限制，0 表示不限制';
COMMENT ON COLUMN api_keys.daily_request_limit IS '每日请求数限制，0 表示不限制';
COMMENT ON COLUMN api_keys.daily_token_limit IS '每日 Token 限制，0 表示不限制';
COMMENT ON COLUMN api_keys.expires_at IS '过期时间，为空表示不过期';
COMMENT ON COLUMN api_keys.last_used_at IS '最近成功使用时间';
COMMENT ON COLUMN api_keys.last_error_at IS '最近调用错误时间';
COMMENT ON COLUMN api_keys.last_error_message IS '最近调用错误信息';
COMMENT ON COLUMN api_keys.created_at IS '创建时间';
COMMENT ON COLUMN api_keys.updated_at IS '更新时间';

CREATE TABLE providers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    slug VARCHAR(80) NOT NULL UNIQUE,
    vendor VARCHAR(40) NOT NULL,
    adapter_type VARCHAR(40) NOT NULL,
    auth_type VARCHAR(40) NOT NULL DEFAULT 'api_key',
    base_url TEXT NOT NULL DEFAULT '',
    api_key_encrypted TEXT NOT NULL DEFAULT '',
    config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_providers_vendor CHECK (vendor IN ('openai', 'anthropic', 'google', 'custom')),
    CONSTRAINT chk_providers_adapter_type CHECK (adapter_type IN ('openai_compatible', 'anthropic', 'claude_code', 'gemini', 'vertexai')),
    CONSTRAINT chk_providers_auth_type CHECK (auth_type IN ('api_key', 'local_oauth', 'adc', 'none')),
    CONSTRAINT chk_providers_status CHECK (status IN ('active', 'disabled'))
);

CREATE INDEX idx_providers_vendor ON providers(vendor);
CREATE INDEX idx_providers_adapter_type ON providers(adapter_type);
CREATE INDEX idx_providers_status ON providers(status);

CREATE TRIGGER trg_providers_updated_at
BEFORE UPDATE ON providers
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE providers IS '上游供应商或上游通道，例如 Anthropic 官方、Claude Code Max、Google Vertex AI';
COMMENT ON COLUMN providers.id IS '供应商主键 ID';
COMMENT ON COLUMN providers.name IS '供应商显示名称';
COMMENT ON COLUMN providers.slug IS '供应商唯一标识';
COMMENT ON COLUMN providers.vendor IS '厂商归属：openai、anthropic、google 或 custom';
COMMENT ON COLUMN providers.adapter_type IS '协议适配器类型：openai_compatible、anthropic、claude_code、gemini 或 vertexai';
COMMENT ON COLUMN providers.auth_type IS '鉴权方式：api_key、local_oauth、adc 或 none';
COMMENT ON COLUMN providers.base_url IS '上游基础地址，不包含具体接口路径';
COMMENT ON COLUMN providers.api_key_encrypted IS '加密后的上游 API Key；local_oauth、adc、none 可为空';
COMMENT ON COLUMN providers.config_json IS '供应商附加配置 JSON，例如版本号、区域、项目 ID';
COMMENT ON COLUMN providers.status IS '供应商状态：active 启用，disabled 禁用';
COMMENT ON COLUMN providers.description IS '供应商备注说明';
COMMENT ON COLUMN providers.created_at IS '创建时间';
COMMENT ON COLUMN providers.updated_at IS '更新时间';

CREATE TABLE model_families (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(80) NOT NULL UNIQUE,
    display_name VARCHAR(120) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_model_families_status CHECK (status IN ('active', 'disabled'))
);

CREATE INDEX idx_model_families_status ON model_families(status);

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

CREATE TABLE models (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL UNIQUE,
    display_name VARCHAR(160) NOT NULL DEFAULT '',
    family VARCHAR(80) NOT NULL DEFAULT '',
    modality VARCHAR(40) NOT NULL DEFAULT 'text',
    status VARCHAR(20) NOT NULL DEFAULT 'enabled',
    active_provider_model_id BIGINT NULL,
    description TEXT NOT NULL DEFAULT '',
    pricing_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_models_modality CHECK (modality IN ('text', 'vision', 'embedding', 'multimodal')),
    CONSTRAINT chk_models_status CHECK (status IN ('enabled', 'disabled'))
);

CREATE INDEX idx_models_status ON models(status);
CREATE INDEX idx_models_family ON models(family);

CREATE TRIGGER trg_models_updated_at
BEFORE UPDATE ON models
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE models IS '对外提供给用户调用的模型目录';
COMMENT ON COLUMN models.id IS '模型主键 ID';
COMMENT ON COLUMN models.name IS '对外模型名称，用户请求中的 model 值';
COMMENT ON COLUMN models.display_name IS '模型展示名称';
COMMENT ON COLUMN models.family IS '模型族或分类，例如 claude、gpt、gemini';
COMMENT ON COLUMN models.modality IS '模型能力类型：text、vision、embedding 或 multimodal';
COMMENT ON COLUMN models.status IS '模型状态：enabled 启用，disabled 禁用';
COMMENT ON COLUMN models.active_provider_model_id IS '当前手动选中的供应商模型路由 ID';
COMMENT ON COLUMN models.description IS '模型备注说明';
COMMENT ON COLUMN models.pricing_json IS '平台模型计价规则 JSON，用于用户侧用量统计和结算';
COMMENT ON COLUMN models.config_json IS '模型附加配置 JSON';
COMMENT ON COLUMN models.created_at IS '创建时间';
COMMENT ON COLUMN models.updated_at IS '更新时间';

CREATE TABLE provider_models (
    id BIGSERIAL PRIMARY KEY,
    provider_id BIGINT NOT NULL REFERENCES providers(id) ON DELETE RESTRICT,
    model_id BIGINT NULL REFERENCES models(id) ON DELETE RESTRICT,
    upstream_model VARCHAR(180) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'enabled',
    max_tokens INTEGER NOT NULL DEFAULT 0,
    timeout_seconds INTEGER NOT NULL DEFAULT 300,
    input_cost_per_1m NUMERIC(12, 6) NOT NULL DEFAULT 0,
    output_cost_per_1m NUMERIC(12, 6) NOT NULL DEFAULT 0,
    pricing_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_provider_models_status CHECK (status IN ('enabled', 'disabled')),
    CONSTRAINT chk_provider_models_max_tokens CHECK (max_tokens >= 0),
    CONSTRAINT chk_provider_models_timeout_seconds CHECK (timeout_seconds > 0),
    CONSTRAINT chk_provider_models_input_cost CHECK (input_cost_per_1m >= 0),
    CONSTRAINT chk_provider_models_output_cost CHECK (output_cost_per_1m >= 0),
    CONSTRAINT uq_provider_models_route UNIQUE (provider_id, model_id, upstream_model)
);

CREATE INDEX idx_provider_models_provider_id ON provider_models(provider_id);
CREATE INDEX idx_provider_models_model_id ON provider_models(model_id);
CREATE INDEX idx_provider_models_status ON provider_models(status);

CREATE TRIGGER trg_provider_models_updated_at
BEFORE UPDATE ON provider_models
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMENT ON TABLE provider_models IS '供应商可提供的具体模型路径，连接对外模型与上游真实模型';
COMMENT ON COLUMN provider_models.id IS '供应商模型主键 ID';
COMMENT ON COLUMN provider_models.provider_id IS '所属供应商 ID';
COMMENT ON COLUMN provider_models.model_id IS '关联的平台模型 ID；为空表示供应商已提供该上游模型，但平台暂未对外开放';
COMMENT ON COLUMN provider_models.upstream_model IS '上游真实模型名称';
COMMENT ON COLUMN provider_models.status IS '供应商模型状态：enabled 启用，disabled 禁用';
COMMENT ON COLUMN provider_models.max_tokens IS '默认最大输出 Token，0 表示不覆盖请求';
COMMENT ON COLUMN provider_models.timeout_seconds IS '该上游模型请求超时时间，单位秒';
COMMENT ON COLUMN provider_models.input_cost_per_1m IS '每百万输入 Token 成本';
COMMENT ON COLUMN provider_models.output_cost_per_1m IS '每百万输出 Token 成本';
COMMENT ON COLUMN provider_models.pricing_json IS '供应商模型成本规则 JSON，用于上游成本统计';
COMMENT ON COLUMN provider_models.config_json IS '供应商模型附加配置 JSON';
COMMENT ON COLUMN provider_models.created_at IS '创建时间';
COMMENT ON COLUMN provider_models.updated_at IS '更新时间';

ALTER TABLE models
    ADD CONSTRAINT fk_models_active_provider_model
        FOREIGN KEY (active_provider_model_id)
        REFERENCES provider_models(id)
        ON DELETE SET NULL;

CREATE TABLE request_logs (
    id BIGSERIAL PRIMARY KEY,
    request_id UUID NOT NULL DEFAULT gen_random_uuid(),
    trace_id VARCHAR(80) NOT NULL DEFAULT '',
    user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    api_key_id BIGINT NULL REFERENCES api_keys(id) ON DELETE SET NULL,
    model_id BIGINT NULL REFERENCES models(id) ON DELETE SET NULL,
    public_model_name VARCHAR(120) NOT NULL DEFAULT '',
    provider_id BIGINT NULL REFERENCES providers(id) ON DELETE SET NULL,
    provider_model_id BIGINT NULL REFERENCES provider_models(id) ON DELETE SET NULL,
    adapter_type VARCHAR(40) NOT NULL DEFAULT '',
    upstream_model VARCHAR(180) NOT NULL DEFAULT '',
    request_type VARCHAR(40) NOT NULL DEFAULT '',
    stream BOOLEAN NOT NULL DEFAULT FALSE,
    client_ip VARCHAR(80) NOT NULL DEFAULT '',
    request_method VARCHAR(12) NOT NULL DEFAULT '',
    request_path TEXT NOT NULL DEFAULT '',
    http_status INTEGER NOT NULL DEFAULT 0,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    estimated_cost NUMERIC(14, 8) NOT NULL DEFAULT 0,
    error_type VARCHAR(80) NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    request_preview JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_preview JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_request_logs_http_status CHECK (http_status >= 0),
    CONSTRAINT chk_request_logs_latency CHECK (latency_ms >= 0),
    CONSTRAINT chk_request_logs_prompt_tokens CHECK (prompt_tokens >= 0),
    CONSTRAINT chk_request_logs_completion_tokens CHECK (completion_tokens >= 0),
    CONSTRAINT chk_request_logs_total_tokens CHECK (total_tokens >= 0),
    CONSTRAINT chk_request_logs_estimated_cost CHECK (estimated_cost >= 0)
);

CREATE UNIQUE INDEX idx_request_logs_request_id ON request_logs(request_id);
CREATE INDEX idx_request_logs_created_at ON request_logs(created_at DESC);
CREATE INDEX idx_request_logs_user_created_at ON request_logs(user_id, created_at DESC);
CREATE INDEX idx_request_logs_api_key_created_at ON request_logs(api_key_id, created_at DESC);
CREATE INDEX idx_request_logs_model_created_at ON request_logs(model_id, created_at DESC);
CREATE INDEX idx_request_logs_provider_created_at ON request_logs(provider_id, created_at DESC);
CREATE INDEX idx_request_logs_provider_model_created_at ON request_logs(provider_model_id, created_at DESC);
CREATE INDEX idx_request_logs_success_created_at ON request_logs(success, created_at DESC);
CREATE INDEX idx_request_logs_error_type_created_at ON request_logs(error_type, created_at DESC);

COMMENT ON TABLE request_logs IS '请求日志和用量统计事实表';
COMMENT ON COLUMN request_logs.id IS '请求日志主键 ID';
COMMENT ON COLUMN request_logs.request_id IS '请求唯一 ID';
COMMENT ON COLUMN request_logs.trace_id IS '链路追踪 ID';
COMMENT ON COLUMN request_logs.user_id IS '发起请求的用户 ID';
COMMENT ON COLUMN request_logs.api_key_id IS '发起请求的 API Key ID';
COMMENT ON COLUMN request_logs.model_id IS '对外模型 ID';
COMMENT ON COLUMN request_logs.public_model_name IS '请求中的对外模型名称快照';
COMMENT ON COLUMN request_logs.provider_id IS '实际命中的供应商 ID';
COMMENT ON COLUMN request_logs.provider_model_id IS '实际命中的供应商模型路由 ID';
COMMENT ON COLUMN request_logs.adapter_type IS '实际使用的协议适配器类型';
COMMENT ON COLUMN request_logs.upstream_model IS '实际调用的上游模型名称';
COMMENT ON COLUMN request_logs.request_type IS '请求类型，例如 chat_completions、messages、embeddings';
COMMENT ON COLUMN request_logs.stream IS '是否为流式请求';
COMMENT ON COLUMN request_logs.client_ip IS '客户端 IP';
COMMENT ON COLUMN request_logs.request_method IS 'HTTP 请求方法';
COMMENT ON COLUMN request_logs.request_path IS '网关请求路径';
COMMENT ON COLUMN request_logs.http_status IS '网关返回 HTTP 状态码';
COMMENT ON COLUMN request_logs.success IS '请求是否成功';
COMMENT ON COLUMN request_logs.latency_ms IS '请求总耗时，单位毫秒';
COMMENT ON COLUMN request_logs.prompt_tokens IS '输入 Token 数';
COMMENT ON COLUMN request_logs.completion_tokens IS '输出 Token 数';
COMMENT ON COLUMN request_logs.total_tokens IS '总 Token 数';
COMMENT ON COLUMN request_logs.estimated_cost IS '按供应商模型价格估算的成本';
COMMENT ON COLUMN request_logs.error_type IS '归一化错误类型';
COMMENT ON COLUMN request_logs.error_message IS '错误信息';
COMMENT ON COLUMN request_logs.request_preview IS '请求内容预览，避免保存过大或敏感完整载荷';
COMMENT ON COLUMN request_logs.response_preview IS '响应内容预览，避免保存过大或敏感完整载荷';
COMMENT ON COLUMN request_logs.metadata IS '附加结构化元数据';
COMMENT ON COLUMN request_logs.created_at IS '创建时间';
