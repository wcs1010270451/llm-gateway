-- 放开 providers.auth_type 约束，新增 claude_oauth：
-- 网关原生用 owner 自有 Claude Max 订阅直连 Anthropic 时使用该鉴权方式。
ALTER TABLE providers
    DROP CONSTRAINT IF EXISTS chk_providers_auth_type;

ALTER TABLE providers
    ADD CONSTRAINT chk_providers_auth_type
        CHECK (auth_type IN ('api_key', 'local_oauth', 'adc', 'none', 'claude_oauth'));

COMMENT ON COLUMN providers.auth_type IS '鉴权方式：api_key、local_oauth、adc、none 或 claude_oauth';
