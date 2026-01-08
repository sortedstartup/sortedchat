# Agent Service Database Schema

This document outlines the database schema design for the Agent Service, supporting both SQLite and PostgreSQL. The design prioritizes flexibility for the evolving nature of LLM agents, enabling features like long-running "experience" sessions, complex tool interactions, and future grading/artifact capabilities.

## 1. Agents Table (`agents`)

Stores the configuration and identity of AI agents. Use this as a catalog of available "employees" or bots.

### Schema (~SQL)
```sql
CREATE TABLE agents (
    id TEXT PRIMARY KEY,                  -- UUID
    name TEXT NOT NULL,                   -- Display name
    description TEXT,                     -- Human readable description
    system_prompt TEXT,                   -- The core personality/instruction
    provider TEXT NOT NULL,               -- e.g., 'openai', 'anthropic'
    model TEXT NOT NULL,                  -- e.g., 'gpt-4', 'claude-3-opus'
    
    -- Configuration
    local_tools TEXT,                     -- JSON Array of tool names: ["calculator", "search"]
                                          -- SQLite: store as TEXT. Postgres: JSONB.
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Future Proofing
    -- config TEXT                        -- JSON for provider specific params (temp, max_tokens, top_p)
    -- is_public BOOLEAN DEFAULT FALSE    -- For sharing agents across projects/users
    -- owner_id TEXT                      -- Link to user/organization table
    -- avatar_url TEXT                    -- UI representation
);
```

### Rationale & Future Thoughts
- **`local_tools` as JSON**: Tools are a list that changes. Storing as a JSON array prevents the need for a many-to-many join table for simple configurations, making reads faster.
- **`system_prompt`**: This is the "soul" of the agent. In the future, this might become a reference to a `prompts` table if we want versioned prompts.
- **Provider/Model separation**: Allows an agent to switch brains (e.g. upgrade from GPT-3.5 to GPT-4) without changing the agent's identity (`id`).

---

## 2. Agent Sessions Table (`agent_sessions`)

A session represents a thread of conversation or a context. As per requirements, this entities handles "Task Sessions", "Experience", and "Company Knowledge".

### Schema (~SQL)
```sql
CREATE TABLE agent_sessions (
    id TEXT PRIMARY KEY,                  -- UUID
    agent_id TEXT NOT NULL,               -- FK -> agents.id
    user_id TEXT,                         -- owner of the session (can be null for system sessions)
    
    -- Session classification
    session_type TEXT DEFAULT 'task',     -- 'task', 'experience', 'company_knowledge'
    
    -- State Management
    status TEXT DEFAULT 'active',         -- 'active', 'archived', 'paused', 'waiting_for_input'
    title TEXT,                           -- Auto-generated or user-set title
    
    -- Context / History (The "Tree" of sessions)
    parent_session_id TEXT,               -- FK -> agent_sessions.id
                                          -- Allows for branching or "forking" a task from an "experience" session.
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_interaction_at TIMESTAMP
    
    -- Future Proofing
    -- context_snapshot TEXT              -- JSON blob of the agent's memory/variable state for easy resumption
    -- summary TEXT                       -- Rolling summary of the session to reduce token usage on load
    -- cost_incurred REAL                 -- Track spend per session
);
```

### Rationale & Future Thoughts
- **`session_type`**: Crucial for the requirement.
    - `experience`: A never-ending session where the agent "lives".
    - `company_knowledge`: A shared read-heavy session (or RAG source).
    - `task`: Ephemeral, goal-oriented sessions.
- **`parent_session_id`**: This supports the "Clone State" or "Fork" pattern. If an agent has a "base experience", a new task can start as a child of that session, inheriting its context (in logic) without polluting the parent.
- **`status` ('waiting_for_input')**: Explicitly modeling the "waiting for user" state helps the UI know when to prompt the user, solving the "how to wait for input" problem significantly.

---

## 3. Agent Messages Table (`agent_messages`)

Stores the actual interaction log. This table is designed to support the granular nature of "Streaming", "Thinking", and "Tool Usage".

### Schema (~SQL)
```sql
CREATE TABLE agent_messages (
    id TEXT PRIMARY KEY,                  -- UUID
    session_id TEXT NOT NULL,             -- FK -> agent_sessions.id
    sequence_number INTEGER NOT NULL,     -- Enforces ordering within a session
    
    -- Classification
    role TEXT NOT NULL,                   -- 'user', 'assistant', 'system', 'tool'
    type TEXT NOT NULL,                   -- 'text', 'thinking', 'tool_call', 'tool_result', 'error'
    
    -- Content Payload
    content TEXT,                         -- Main text payload (message text, thought chain, or error message)
    
    -- Tool Specifics (populated when type = 'tool_call' or 'tool_result')
    tool_name TEXT,                       -- e.g., 'google_search', 'get_user_input'
    tool_call_id TEXT,                    -- Unique ID from the LLM to correlate call <-> result
    tool_args TEXT,                       -- JSON string of arguments provided by LLM
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Future Proofing
    -- token_count INTEGER                -- Usage tracking per message
    -- is_hidden BOOLEAN DEFAULT FALSE    -- Hide 'thinking' or specific tool steps from end-users?
    -- feedback_score INTEGER             -- -1 (thumbs down) to 1 (thumbs up) for specific message reinforcement
);
```

### Rationale & Future Thoughts
- **Granular Types (`thinking`, `tool_call`)**: Instead of squashing everything into one "Assistant" message, we treat Thoughts and Tool Calls as first-class events. This allows:
    - Replaying the *exact* stream in the UI.
    - Analyzing how much time/tokens the agent spends "Thinking" vs "Doing".
    - Debugging tool loops easily.
- **`tool_call_id`**: Essential for modern LLM tool usage (e.g., OpenAI Tool API requires matching IDs).
- **Stream Friendly**: This schema maps 1:1 with the `ChatResponse` oneof fields in `chatservice.proto`.

---

<!-- ## 4. Agent Artifacts Table (Proposed Future)

User asked: *"how do agent produce artifacts ? how to show ? store them ?"*

Agents often produce code, documents, or images that shouldn't just be bury in the chat log.

### Schema (Draft)
```sql
CREATE TABLE agent_artifacts (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,             -- FK -> agent_sessions.id
    message_id TEXT NOT NULL,             -- FK -> agent_messages.id (Which message created this)
    
    name TEXT,                            -- Filename or Title (e.g., "snake_game.py")
    type TEXT NOT NULL,                   -- 'code', 'markdown', 'svg', 'image'
    content TEXT,                         -- The actual content (or blob reference)
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
``` -->
<!-- 
### Rationale
- Decouples "Output" from "Conversation".
- Allows a "Files" tab in the UI where users can see all generated scripts/docs from a session.

---

## 5. Agent Run Grading (Proposed Future)

User asked: *"grade an agent, auto grade runs"*

To evaluate agent performance over time.

### Schema (Draft)
```sql
CREATE TABLE session_grades (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,             -- FK -> agent_sessions.id
    
    score INTEGER,                        -- 1-10 or 1-5
    feedback TEXT,                        -- Human written feedback
    auto_graded BOOLEAN DEFAULT FALSE,    -- Was this graded by another LLM?
    grader_model TEXT,                    -- If auto-graded, which model did it?
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
``` -->
