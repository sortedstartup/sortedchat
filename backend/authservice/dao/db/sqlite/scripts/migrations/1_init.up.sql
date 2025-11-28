-- SQLite complete schema migration

CREATE TABLE IF NOT EXISTS userservice_users (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    roles TEXT NOT NULL,
    oauth_provider TEXT NOT NULL,
    oauth_user_id TEXT NOT NULL,
    is_federated INTEGER NOT NULL CHECK (is_federated IN (0, 1))
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_userservice_users_email ON userservice_users(email);
CREATE INDEX IF NOT EXISTS idx_userservice_users_oauth_provider_user_id ON userservice_users(oauth_provider, oauth_user_id);