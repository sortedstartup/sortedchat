ALTER TABLE agent_messages ADD COLUMN success BOOLEAN DEFAULT TRUE;
ALTER TABLE agent_messages ADD COLUMN error_message TEXT;
ALTER TABLE agent_messages ADD COLUMN run_time_ms BIGINT DEFAULT 0;
