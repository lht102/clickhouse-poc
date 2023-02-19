// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.0

package db

import (
	"database/sql"
	"time"
)

type Category struct {
	ID         int64
	EnName     string
	CreateTime time.Time
	UpdateTime time.Time
}

type Event struct {
	ID         int64
	EnName     string
	StartTime  time.Time
	EndTime    sql.NullTime
	CategoryID int64
	CreateTime time.Time
	UpdateTime time.Time
}
