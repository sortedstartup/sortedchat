CREATE TABLE IF NOT EXISTS rag_chunks (
    id TEXT PRIMARY KEY NOT NULL,
    project_id TEXT NOT NULL,
    docs_id TEXT NOT NULL,
    source TEXT,
    start_byte INTEGER,
    end_byte INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- CREATE VIRTUAL TABLE IF NOT EXISTS rag_chunks_vec USING vec0(
-- 	id INTEGER,
--     embedding FLOAT[1024]
-- );