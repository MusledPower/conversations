CREATE TABLE IF NOT EXISTS rooms (
    id UUID primary key,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(255),
    capacity INTEGER,
    created_at timestamp
);