-- Migration: Add content_json column to messages table for multi-modal content support
ALTER TABLE chat_messages ADD COLUMN content_json TEXT;

-- Keep 'content' column for backward compatibility and text-only display
-- content_json will be NULL for old messages and text-only new messages
