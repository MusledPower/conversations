CREATE TABLE IF NOT EXISTS slots (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL,
    start_time timestamp,
    end_time timestamp,

    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
);

