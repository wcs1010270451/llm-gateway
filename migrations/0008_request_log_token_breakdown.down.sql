ALTER TABLE request_logs
    DROP COLUMN IF EXISTS tool_tokens,
    DROP COLUMN IF EXISTS reasoning_tokens,
    DROP COLUMN IF EXISTS cache_read_input_tokens,
    DROP COLUMN IF EXISTS cache_creation_input_tokens;
