ALTER TABLE chat_messages ADD COLUMN type TEXT NOT NULL DEFAULT 'text';
ALTER TABLE chat_messages ADD COLUMN tool_name TEXT;
ALTER TABLE chat_messages ADD COLUMN tool_call_id TEXT;
ALTER TABLE chat_messages ADD COLUMN tool_args TEXT;
