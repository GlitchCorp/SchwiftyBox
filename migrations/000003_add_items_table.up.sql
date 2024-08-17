CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    backpack_id VARCHAR(20) NOT NULL,
    description VARCHAR(1000),
    added_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_email VARCHAR(255) REFERENCES users(email) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES items(id) ON DELETE SET NULL
);

-- Create index for faster lookups
CREATE INDEX idx_items_user_email ON items(user_email);
CREATE INDEX idx_items_parent_id ON items(parent_id);
CREATE INDEX idx_items_backpack_id ON items(backpack_id);
