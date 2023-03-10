// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0

package db

import (
	"database/sql"
	"time"
)

type Payment struct {
	ID            int64
	PaymentStatus string
	CompleteTime  sql.NullTime
	CreateTime    time.Time
	UpdateTime    time.Time
}

type Transaction struct {
	ID                int64
	TransactionStatus string
	EventID           int64
	PaymentID         int64
	CreateTime        time.Time
	UpdateTime        time.Time
}
