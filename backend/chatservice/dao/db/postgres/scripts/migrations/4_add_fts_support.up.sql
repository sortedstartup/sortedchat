-- Migration: 4_add_fts_support.up.sql
-- Add full text search support to chat_messages table

-- Add tsvector column for full text search
ALTER TABLE chat_messages ADD COLUMN content_tsvector tsvector;

-- Populate existing data with tsvector
UPDATE chat_messages SET content_tsvector = to_tsvector('english', content);

-- Create GIN index for fast FTS
CREATE INDEX idx_chat_messages_fts 
    ON chat_messages USING gin (content_tsvector)
    WITH (fastupdate = off);  -- Optimize for read performance

-- Create function to update tsvector when content changes
CREATE OR REPLACE FUNCTION update_content_tsvector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.content_tsvector = to_tsvector('english', NEW.content);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to keep tsvector updated automatically
CREATE TRIGGER chat_messages_tsvector_update
    BEFORE INSERT OR UPDATE OF content ON chat_messages
    FOR EACH ROW EXECUTE FUNCTION update_content_tsvector();
