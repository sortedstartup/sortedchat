ALTER TABLE model_metadata ADD COLUMN cached_token_cost REAL;

ALTER TABLE chat_messages ADD COLUMN cached_token_count INTEGER DEFAULT 0;
ALTER TABLE chat_messages ADD COLUMN cost REAL DEFAULT 0;


ALTER TABLE chat_list ADD COLUMN cost REAL DEFAULT 0;
ALTER TABLE chat_list ADD COLUMN input_token_count INTEGER DEFAULT 0;
ALTER TABLE chat_list ADD COLUMN output_token_count INTEGER DEFAULT 0;
ALTER TABLE chat_list ADD COLUMN cached_token_count INTEGER DEFAULT 0;