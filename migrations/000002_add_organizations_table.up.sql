CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add organization_users junction table
CREATE TABLE organization_users (
    organization_id INTEGER REFERENCES organizations(id) ON DELETE CASCADE,
    user_email VARCHAR(255) REFERENCES users(email) ON DELETE CASCADE,
    PRIMARY KEY (organization_id, user_email)
);

-- Add active_organization_id to users table
ALTER TABLE users ADD COLUMN active_organization_id INTEGER REFERENCES organizations(id);
ALTER TABLE users ADD COLUMN prefix VARCHAR(10);
