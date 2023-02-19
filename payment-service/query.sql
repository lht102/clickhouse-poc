-- name: CreatePayment :execresult
INSERT INTO payments (
    id, payment_status, complete_time,
    create_time, update_time
)
VALUES (?, ?, ?, ?, ?);

-- name: CreateTransaction :execresult
INSERT INTO transactions (
    transaction_status, event_id,
    payment_id, create_time, update_time
)
VALUES (?, ?, ?, ?, ?);
