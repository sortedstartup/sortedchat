-- Add user_id column to main tables
ALTER TABLE chat_list ADD COLUMN user_id TEXT DEFAULT '0' NOT NULL;
ALTER TABLE chat_messages ADD COLUMN user_id TEXT DEFAULT '0' NOT NULL;
ALTER TABLE project ADD COLUMN user_id TEXT DEFAULT '0' NOT NULL;
ALTER TABLE project_docs ADD COLUMN user_id TEXT DEFAULT '0' NOT NULL;
ALTER TABLE rag_chunks ADD COLUMN user_id TEXT DEFAULT '0' NOT NULL;

-- Add indexes for performance
CREATE INDEX idx_chat_list_user_id ON chat_list(user_id);
CREATE INDEX idx_chat_messages_user_id ON chat_messages(user_id);
CREATE INDEX idx_project_user_id ON project(user_id);
CREATE INDEX idx_project_docs_user_id ON project_docs(user_id);
CREATE INDEX idx_rag_chunks_user_id ON rag_chunks(user_id);

-- Compound indexes for common queries
CREATE INDEX idx_chat_list_user_project ON chat_list(user_id, project_id);
CREATE INDEX idx_chat_messages_user_chat ON chat_messages(user_id, chat_id);
