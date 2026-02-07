-- Add mcp_servers column to agents table
ALTER TABLE agents ADD COLUMN mcp_servers TEXT DEFAULT '[]';

