-- Migration: 3_add_vector_column.up.sql
-- Add vector column to rag_chunks table for Phase 2 vector operations

-- Add vector column for embeddings (768 dimension as per CHOICE 2 in the document)
ALTER TABLE rag_chunks ADD COLUMN embedding vector(768);

-- Create HNSW index for efficient similarity search using cosine distance
CREATE INDEX idx_rag_chunks_embedding_hnsw 
    ON rag_chunks USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- Add metadata for vector operations
ALTER TABLE rag_chunks ADD COLUMN embedding_model TEXT DEFAULT 'text-embedding-3-small';
ALTER TABLE rag_chunks ADD COLUMN embedding_created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
