ALTER TABLE chat_list ADD COLUMN parent_chat_id TEXT;
ALTER TABLE chat_list ADD COLUMN parent_message_id TEXT;
ALTER TABLE chat_list ADD COLUMN is_main_branch BOOLEAN DEFAULT TRUE;