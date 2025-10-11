-- Add content_image column to store base64 image data separately
-- This prevents tsvector 1MB limit issues when messages contain large images
ALTER TABLE chat_messages ADD COLUMN content_image TEXT;
