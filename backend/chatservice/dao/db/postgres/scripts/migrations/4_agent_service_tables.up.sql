-- Agent service tables for PostgreSQL
CREATE TABLE IF NOT EXISTS agents (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    system_prompt TEXT,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    local_tools TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS agent_sessions (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    title TEXT,
    parent_session_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(agent_id) REFERENCES agents(id),
    FOREIGN KEY(parent_session_id) REFERENCES agent_sessions(id)
);

CREATE TABLE IF NOT EXISTS agent_messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    sequence_number INTEGER NOT NULL,
    role TEXT NOT NULL,
    type TEXT NOT NULL,
    content TEXT,
    tool_name TEXT,
    tool_call_id TEXT,
    tool_args TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(session_id) REFERENCES agent_sessions(id)
);

CREATE TABLE IF NOT EXISTS agent_docs (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    docs_id TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    uploaded_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_agent_docs_agent_id ON agent_docs(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_docs_docs_id ON agent_docs(docs_id);
CREATE INDEX IF NOT EXISTS idx_agent_docs_path ON agent_docs(agent_id, file_path);

