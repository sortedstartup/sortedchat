 CREATE TABLE IF NOT EXISTS chat_messages (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        chat_id TEXT NOT NULL,
        role TEXT NOT NULL,
        content TEXT NOT NULL,
		model TEXT ,
		error BOOLEAN,
		input_token_count INTEGER,
		output_token_count INTEGER,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS chat_list (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id TEXT NOT NULL,
    project_id TEXT,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS model_metadata (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    provider TEXT,
    input_token_cost REAL,
    output_token_cost REAL
);
CREATE VIRTUAL TABLE IF NOT EXISTS chat_messages_fts USING fts5(
    message_content,
    content='chat_messages',
    content_rowid='id',
    tokenize='porter unicode61'
);

CREATE TRIGGER IF NOT EXISTS chat_messages_ai_fts 
AFTER INSERT ON chat_messages 
BEGIN
INSERT INTO chat_messages_fts (rowid, message_content) VALUES (new.id, new.content);
END;

CREATE TRIGGER IF NOT EXISTS chat_messages_ad_fts 
AFTER DELETE ON chat_messages
BEGIN
INSERT INTO chat_messages_fts (chat_messages_fts, rowid, message_content) VALUES ('delete', old.id, old.content);
END;

CREATE TRIGGER IF NOT EXISTS chat_messages_au_fts 
AFTER UPDATE OF content ON chat_messages 
BEGIN
INSERT INTO chat_messages_fts (chat_messages_fts, rowid, message_content) VALUES ('delete', old.id, old.content);
INSERT INTO chat_messages_fts (rowid, message_content) VALUES (new.id, new.content);
END;