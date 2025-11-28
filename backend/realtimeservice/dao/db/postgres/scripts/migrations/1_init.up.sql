CREATE TABLE IF NOT EXISTS realtimeservice_audio_chat (
    id TEXT PRIMARY KEY,
    model_name TEXT NOT NULL,
    user_id TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL
);