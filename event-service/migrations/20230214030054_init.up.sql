CREATE TABLE categories (
    id bigint NOT NULL AUTO_INCREMENT,
    en_name varchar(255) NOT NULL,
    create_time DATETIME(6) NOT NULL,
    update_time DATETIME(6) NOT NULL,
    PRIMARY KEY(id)
);

CREATE INDEX index_categories_on_update_time ON categories (update_time);

CREATE TABLE events (
    id bigint NOT NULL AUTO_INCREMENT,
    en_name varchar(255) NOT NULL,
    start_time DATETIME(6) NOT NULL,
    end_time DATETIME(6) NULL,
    category_id bigint NOT NULL,
    create_time DATETIME(6) NOT NULL,
    update_time DATETIME(6) NOT NULL,
    PRIMARY KEY(id),
    CONSTRAINT fk_events_category_id_categories FOREIGN KEY (
        category_id
    ) REFERENCES categories (id)
);

CREATE INDEX index_events_on_update_time ON events (update_time);
