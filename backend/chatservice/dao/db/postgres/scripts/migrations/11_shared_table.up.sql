ALTER TABLE model_metadata RENAME TO shared_models_metadata;

ALTER TABLE shared_models_metadata ADD COLUMN progress TEXT;
ALTER TABLE shared_models_metadata ADD COLUMN is_downloaded BOOLEAN;
ALTER TABLE shared_models_metadata ADD COLUMN is_downloadable BOOLEAN;
ALTER TABLE shared_models_metadata ADD COLUMN status INTEGER DEFAULT 0;
ALTER TABLE shared_models_metadata ADD COLUMN filestore_id TEXT;
ALTER TABLE shared_models_metadata ADD COLUMN is_enabled BOOLEAN;
ALTER TABLE shared_models_metadata ADD COLUMN is_embedding_model BOOLEAN;

-- Status values: 0=none, 1=pending, 2=downloading, 3=completed, 4=failed
