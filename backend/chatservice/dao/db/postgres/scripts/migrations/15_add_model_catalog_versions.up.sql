ALTER TABLE shared_models_metadata
ADD COLUMN IF NOT EXISTS json_schema_version TEXT DEFAULT '';

ALTER TABLE shared_models_metadata
ADD COLUMN IF NOT EXISTS model_revision_version TEXT DEFAULT '';
