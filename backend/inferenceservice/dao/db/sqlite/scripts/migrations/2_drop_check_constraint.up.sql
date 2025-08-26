-- Step 1: Create a new table without the CHECK constraint
CREATE TABLE IF NOT EXISTS inferenceservice_models_metadata_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    provider TEXT,
    input_token_cost REAL,
    output_token_cost REAL,
    progress TEXT,
    is_downloaded BOOLEAN DEFAULT FALSE,
    is_downloadable BOOLEAN DEFAULT FALSE,
    status INTEGER DEFAULT 0,  -- CHECK constraint removed
    filestore_id TEXT DEFAULT NULL
);

-- Step 2: Copy all data from the old table to the new table
INSERT INTO inferenceservice_models_metadata_new 
SELECT * FROM inferenceservice_models_metadata;

-- Step 3: Drop the old table
DROP TABLE inferenceservice_models_metadata;

-- Step 4: Rename the new table to the original name
ALTER TABLE inferenceservice_models_metadata_new 
RENAME TO inferenceservice_models_metadata;