-- name: CreateCategory :execresult
INSERT INTO categories (
    id, en_name, create_time, update_time
)
VALUES (?, ?, ?, ?);

-- name: CreateEvent :execresult
INSERT INTO events (
    id, en_name, start_time, end_time,
    category_id, create_time, update_time
)
VALUES (?, ?, ?, ?, ?, ?, ?);
