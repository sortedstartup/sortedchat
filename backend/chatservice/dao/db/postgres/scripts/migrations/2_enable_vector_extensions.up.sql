-- Migration: 2_enable_vector_extensions.up.sql
-- Enable required extensions for vector operations and full text search

CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;  -- For trigram similarity
