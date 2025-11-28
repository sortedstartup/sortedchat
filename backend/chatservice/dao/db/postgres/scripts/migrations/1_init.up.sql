-- PostgreSQL complete schema migration
-- This includes all tables from SQLite migrations 1-9 in a single file

-- Chat messages table
CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGSERIAL PRIMARY KEY,
    chat_id TEXT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    model TEXT,
    error BOOLEAN DEFAULT FALSE,
    input_token_count INTEGER,
    output_token_count INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id TEXT NOT NULL DEFAULT '0'
);

-- Chat list table
CREATE TABLE IF NOT EXISTS chat_list (
    id BIGSERIAL PRIMARY KEY,
    chat_id TEXT NOT NULL,
    name TEXT NOT NULL,
    project_id TEXT,
    parent_chat_id TEXT,
    parent_message_id TEXT,
    is_main_branch BOOLEAN DEFAULT TRUE,
    user_id TEXT NOT NULL DEFAULT '0'
);

-- Model metadata table
CREATE TABLE IF NOT EXISTS model_metadata (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    provider TEXT,
    input_token_cost REAL,
    output_token_cost REAL
);

-- Project table
CREATE TABLE IF NOT EXISTS project (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    additional_data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id TEXT NOT NULL DEFAULT '0'
);

-- Project documents table
CREATE TABLE IF NOT EXISTS project_docs (
    id BIGSERIAL PRIMARY KEY,
    project_id TEXT NOT NULL,
    docs_id TEXT,
    file_name TEXT,
    file_size BIGINT NOT NULL,
    embedding_status INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id TEXT NOT NULL DEFAULT '0'
);

-- RAG chunks table (without vector embeddings for Phase 1)
CREATE TABLE IF NOT EXISTS rag_chunks (
    id TEXT PRIMARY KEY NOT NULL,
    project_id TEXT NOT NULL,
    docs_id TEXT NOT NULL,
    start_byte INTEGER,
    end_byte INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id TEXT NOT NULL DEFAULT '0'
);

-- Settings table
CREATE TABLE IF NOT EXISTS settings (
    name TEXT PRIMARY KEY,
    settings TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_chat_list_user_id ON chat_list(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_id ON chat_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_project_user_id ON project(user_id);
CREATE INDEX IF NOT EXISTS idx_project_docs_user_id ON project_docs(user_id);
CREATE INDEX IF NOT EXISTS idx_rag_chunks_user_id ON rag_chunks(user_id);

-- Compound indexes for common queries
CREATE INDEX IF NOT EXISTS idx_chat_list_user_project ON chat_list(user_id, project_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_chat ON chat_messages(user_id, chat_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_id ON chat_messages(chat_id);
CREATE INDEX IF NOT EXISTS idx_project_docs_project_id ON project_docs(project_id);
CREATE INDEX IF NOT EXISTS idx_rag_chunks_project_id ON rag_chunks(project_id);

-- Trigger to update the updated_at timestamp for settings
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER settings_updated_at 
    BEFORE UPDATE ON settings 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for project table
CREATE TRIGGER project_updated_at 
    BEFORE UPDATE ON project 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for project_docs table  
CREATE TRIGGER project_docs_updated_at 
    BEFORE UPDATE ON project_docs 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for rag_chunks table
CREATE TRIGGER rag_chunks_updated_at 
    BEFORE UPDATE ON rag_chunks 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
