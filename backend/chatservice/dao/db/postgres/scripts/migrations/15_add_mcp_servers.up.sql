-- Add mcp_servers column to agents table
ALTER TABLE agents ADD COLUMN IF NOT EXISTS mcp_servers JSONB DEFAULT '[]'::jsonb;

