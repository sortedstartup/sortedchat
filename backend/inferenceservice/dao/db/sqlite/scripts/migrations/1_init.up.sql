CREATE TABLE IF NOT EXISTS inference_model_metadata (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    provider TEXT,
    input_token_cost REAL,
    output_token_cost REAL,
    progress TEXT,
    is_downloaded BOOLEAN DEFAULT FALSE,
    is_downloadable BOOLEAN DEFAULT FALSE,
    status INTEGER DEFAULT 0 CHECK(status IN (0, 1, 2, 3, 4)),
    filestore_id TEXT DEFAULT NULL
);
-- Status values: 0=none, 1=pending, 2=downloading, 3=completed, 4=failed