ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS plain_key TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN api_keys.plain_key IS 'API Key 明文值，仅用于个人部署下方便查看和管理';
