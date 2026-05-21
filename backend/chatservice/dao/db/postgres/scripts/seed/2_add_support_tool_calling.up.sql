UPDATE shared_models_metadata
SET capabilities = jsonb_set(
    CASE
        WHEN capabilities IS NULL OR btrim(capabilities::text, '"') = '' THEN '{}'::jsonb
        ELSE capabilities
    END,
    '{support_tool_calling}',
    CASE
        WHEN id IN ('gemma-3-270m-it-GGUF')
            THEN 'false'::jsonb
        ELSE 'true'::jsonb
    END,
    true
)
WHERE id IN (
    'gpt-4.1',
    'gpt-4o',
    'o3-mini',
    'o4-mini',
    'o3',
    'gemini-2.5-flash',
    'gemini-3-flash-preview',
    'gemini-2.0-flash',
    'gemini-2.5-pro',
    'gemini-3-pro-preview',
    'claude-3.5-haiku',
    'claude-3-7-sonnet',
    'claude-4-sonnet',
    'gpt-5',
    'gpt-5-mini',
    'gpt-5-nano',
    'gpt-realtime',
    'gpt-realtime-mini',
    'gpt-4o-realtime',
    'gpt-4o-mini-realtime',
    'gemma-3-270m-it-GGUF'
);
