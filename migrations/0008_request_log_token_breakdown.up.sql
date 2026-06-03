ALTER TABLE request_logs
    ADD COLUMN cache_creation_input_tokens INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN cache_read_input_tokens INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN reasoning_tokens INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN tool_tokens INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN request_logs.cache_creation_input_tokens IS 'Tokens written into provider prompt cache for this request';
COMMENT ON COLUMN request_logs.cache_read_input_tokens IS 'Tokens read from provider prompt cache for this request';
COMMENT ON COLUMN request_logs.reasoning_tokens IS 'Reasoning or thinking tokens reported by the provider';
COMMENT ON COLUMN request_logs.tool_tokens IS 'Tool-use prompt tokens reported by the provider';
