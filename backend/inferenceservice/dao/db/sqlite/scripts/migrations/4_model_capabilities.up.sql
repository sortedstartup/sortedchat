ALTER TABLE inferenceservice_models_metadata ADD COLUMN capabilities TEXT NOT NULL DEFAULT '{}';
ALTER TABLE inferenceservice_models_metadata ADD COLUMN cached_token_cost REAL NOT NULL DEFAULT 0;
ALTER TABLE inferenceservice_models_metadata ADD COLUMN is_enabled BOOLEAN NOT NULL DEFAULT TRUE;