-- Agent documents table for storing files uploaded to agents
CREATE TABLE agent_docs (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    docs_id TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    uploaded_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

CREATE INDEX idx_agent_docs_agent_id ON agent_docs(agent_id);
CREATE INDEX idx_agent_docs_docs_id ON agent_docs(docs_id);
CREATE INDEX idx_agent_docs_path ON agent_docs(agent_id, file_path);


