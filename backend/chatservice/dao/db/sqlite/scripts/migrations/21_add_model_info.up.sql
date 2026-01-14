-- Add model_info column to shared_models_metadata table
ALTER TABLE shared_models_metadata ADD COLUMN model_info TEXT DEFAULT '{}';

