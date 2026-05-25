ALTER TABLE chat_messages ADD COLUMN brave_search_count INTEGER DEFAULT 0;
ALTER TABLE chat_messages ADD COLUMN scrape_api_usage_time REAL DEFAULT 0;