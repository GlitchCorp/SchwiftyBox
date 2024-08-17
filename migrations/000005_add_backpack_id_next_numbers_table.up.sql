CREATE TABLE backpack_id_next_numbers (
    id SERIAL PRIMARY KEY,
    backpack_id VARCHAR(20) NOT NULL,
    number INTEGER NOT NULL DEFAULT 1
);

-- Create unique index on backpack_id
CREATE UNIQUE INDEX idx_backpack_id_next_numbers_backpack_id ON backpack_id_next_numbers(backpack_id);
