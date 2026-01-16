-- Add model_info column to shared_models_metadata table
ALTER TABLE shared_models_metadata ADD COLUMN model_info TEXT DEFAULT '{}';

-- Add creator_name, modified_by, and description columns
ALTER TABLE shared_models_metadata ADD COLUMN creator_name TEXT DEFAULT '';
ALTER TABLE shared_models_metadata ADD COLUMN modified_by TEXT DEFAULT '';
ALTER TABLE shared_models_metadata ADD COLUMN description TEXT DEFAULT '';

