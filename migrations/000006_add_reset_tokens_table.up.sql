CREATE TABLE reset_tokens (
    id SERIAL PRIMARY KEY,
    token VARCHAR(30) NOT NULL UNIQUE,
    expired_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_email VARCHAR(255) REFERENCES users(email) ON DELETE CASCADE
);

-- Create index for faster lookups
CREATE INDEX idx_reset_tokens_user_email ON reset_tokens(user_email);
CREATE INDEX idx_reset_tokens_token ON reset_tokens(token);
