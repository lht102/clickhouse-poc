# ClickHouseCtl

1. set up event service and payment service for preparing the fake data

2. copy `named_collections_sample.xml` to `named_collections.xml` and update database configuration.
3. start clickhouse server and client
```bash
docker-compose up -d
docker exec -it my-clickhouse-server clickhouse-client
```

```sql
-- connect ClickHouse to MySQL
CREATE DATABASE source_mysql_event_service Engine = MySQL("mysql_event_service");
CREATE DATABASE source_mysql_payment_service Engine = MySQL("mysql_payment_service");

-- connect analytics database and copy data from MySQL
CREATE DATABASE analytics;

CREATE TABLE analytics.events
Engine = ReplacingMergeTree
ORDER BY id
AS SELECT * FROM source_mysql_event_service.events;

CREATE TABLE analytics.categories
Engine = ReplacingMergeTree
ORDER BY id
AS SELECT * FROM source_mysql_event_service.categories;

CREATE TABLE analytics.payments
Engine = ReplacingMergeTree
ORDER BY id
AS SELECT * FROM source_mysql_payment_service.payments;

CREATE TABLE analytics.transactions
Engine = ReplacingMergeTree
ORDER BY id
AS SELECT * FROM source_mysql_payment_service.transactions;

-- create materialized view to aggregate data
CREATE MATERIALIZED VIEW analytics.transactions_mv (
	id Int64,
	category_id Int64,
	settlement_duration_seconds Int64
	) ENGINE ReplacingMergeTree
ORDER BY (id, category_id) AS
SELECT
	t.id,
	category_id,
	(ev.end_time - t.update_time) settlement_duration_seconds
FROM
		 analytics.transactions t
LEFT JOIN analytics.events ev ON
		t.event_id = ev.id 
WHERE
		t.transaction_status = 'SETTLED'
	AND ev.end_time IS NOT NULL;

-- query data for reporting
SELECT
	category_id ,
	quantile(0.95)(diff) AS t95_settlement_duration_seconds
FROM
	(
	SELECT
		category_id,
		(ev.end_time - t.update_time) diff
	FROM
		analytics.events ev
	JOIN analytics.transactions t ON
		ev.id = t.event_id
	WHERE
		t.transaction_status = 'SETTLED'
		AND ev.end_time IS NOT NULL)
GROUP BY
	category_id;

-- query data for using materialized view
SELECT
	category_id ,
	quantile(0.95)(settlement_duration_seconds) AS t95_settlement_duration_seconds
FROM analytics.transactions_mv
GROUP BY
	category_id;

-- incremental update
INSERT
	INTO
	analytics.events
SELECT
	*
FROM
	source_mysql_event_service.events se
WHERE
	se.update_time >= (
	SELECT
		max(ae.update_time)
	FROM
		analytics.events ae);

```
