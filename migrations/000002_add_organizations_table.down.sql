ALTER TABLE users DROP COLUMN IF EXISTS prefix;
ALTER TABLE users DROP COLUMN IF EXISTS active_organization_id;
DROP TABLE IF EXISTS organization_users;
DROP TABLE IF EXISTS organizations;

