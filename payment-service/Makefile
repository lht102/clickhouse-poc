SHELL := /bin/sh

MIGRATION_PATH := ./migrations/

-include .env.$(ENV)
export

.PHONY: run-payment-service
run-payment-service:
	go run cmd/payment-service/main.go

.PHONY: build
build:
	go build -v -o ./bin/payment-service ./cmd/payment-service/*

.PHONY: dep
dep:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test ./... -race -cover -coverprofile=c.out -covermode=atomic

.PHONY: sqlc-gen
sqlc-gen:
	sqlc generate

.PHONY: sql-lint
sql-lint:
	sqlfluff lint ./migrations/ --dialect mysql

.PHONY: migrate-create
migrate-create:
	@migrate create -dir $(MIGRATION_PATH) -ext sql $(name)

.PHONY: migrate
migrate:
	@migrate -path $(MIGRATION_PATH) \
		-database "mysql://$(MYSQL_USERNAME):$(MYSQL_PASSWORD)@$(MYSQL_PROTOCOL)($(MYSQL_ADDRESS))/$(MYSQL_DATABASE)" up

.PHONY: reset-db
reset-db:
	@mysql -h $(MYSQL_ADDRESS) -u $(MYSQL_USERNAME) --password=$(MYSQL_PASSWORD) --protocol=$(MYSQL_PROTOCOL) -e "DROP DATABASE IF EXISTS $(MYSQL_DATABASE); CREATE DATABASE $(MYSQL_DATABASE);"

.PHONY: migrate-reset
migrate-reset: reset-db migrate
