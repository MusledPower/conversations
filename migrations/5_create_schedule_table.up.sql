CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL UNIQUE,
    days_of_week INT[] NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,

    FOREIGN KEY (room_id) REFERENCES rooms (id) ON DELETE CASCADE
);