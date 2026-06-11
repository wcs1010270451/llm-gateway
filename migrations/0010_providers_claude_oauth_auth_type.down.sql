-- 回滚：恢复不含 claude_oauth 的 auth_type 约束。
-- 注意：若已有 auth_type='claude_oauth' 的 provider，需先改掉或删除，否则约束无法重建。
ALTER TABLE providers
    DROP CONSTRAINT IF EXISTS chk_providers_auth_type;

ALTER TABLE providers
    ADD CONSTRAINT chk_providers_auth_type
        CHECK (auth_type IN ('api_key', 'local_oauth', 'adc', 'none'));

COMMENT ON COLUMN providers.auth_type IS '鉴权方式：api_key、local_oauth、adc 或 none';
