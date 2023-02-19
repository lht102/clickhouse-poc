CREATE TABLE payments (
    id bigint NOT NULL AUTO_INCREMENT,
    payment_status varchar(50) NOT NULL, -- ACCEPTED/REJECTED/COMPLETED
    complete_time DATETIME(6) NULL,
    create_time DATETIME(6) NOT NULL,
    update_time DATETIME(6) NOT NULL,
    PRIMARY KEY(id)
);

CREATE INDEX index_payments_on_update_time ON payments (update_time);

CREATE TABLE transactions (
    id bigint NOT NULL AUTO_INCREMENT,
    transaction_status varchar(50) NOT NULL, -- PENDING/SETTLED
    event_id bigint NOT NULL,
    payment_id bigint NOT NULL,
    create_time DATETIME(6) NOT NULL,
    update_time DATETIME(6) NOT NULL,
    PRIMARY KEY(id),
    CONSTRAINT fk_transactions_payment_id_payments FOREIGN KEY (
        payment_id
    ) REFERENCES payments (id)
);

CREATE INDEX index_transactions_on_update_time ON transactions (update_time);
