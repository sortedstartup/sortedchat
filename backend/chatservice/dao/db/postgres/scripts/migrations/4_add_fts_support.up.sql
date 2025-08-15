-- Migration: 4_add_fts_support.up.sql
-- Use modern PostgreSQL generated column syntax (PostgreSQL 12+)
ALTER TABLE chat_messages ADD COLUMN content_tsvector tsvector
    GENERATED ALWAYS AS (to_tsvector('english', content)) STORED;
-- Create GIN index for fast FTS
CREATE INDEX idx_chat_messages_fts
    ON chat_messages USING gin (content_tsvector)
    WITH (fastupdate = off);  -- Optimize for read performance
-- Note: No triggers needed - PostgreSQL automatically maintains the generated column