CREATE TABLE IF NOT EXISTS bookings (
    id UUID primary key,
    slot_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(255) NOT NULL,
    conference_link TEXT,
    CREATED_AT TIMESTAMP,

    FOREIGN KEY (slot_id) REFERENCES slots (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX bookings_one_active_per_slot
    ON bookings(slot_id)
    WHERE status = 'active';