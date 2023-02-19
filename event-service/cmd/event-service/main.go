package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jaswdr/faker"
	"github.com/lht102/clickhouse-poc/event-service/db"
	"github.com/pacedotdev/batch"
	"github.com/urfave/cli/v2"
)

func main() {
	var (
		mysqlUsername     string
		mysqlPassword     string
		mysqlDatabase     string
		mysqlProtocol     string
		mysqlAddress      string
		mysqlMaxOpenConns int
		mysqlMaxIdleConns int

		batchSize                 int
		batchIntervalMilliseconds int
		startingCategoryID        int64
		categoryCount             int
		startingEventID           int64
		eventCount                int
	)

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "mysql-username",
				Value:       "root",
				EnvVars:     []string{"MYSQL_USERNAME"},
				Destination: &mysqlUsername,
			},
			&cli.StringFlag{
				Name:        "mysql-password",
				Value:       "",
				EnvVars:     []string{"MYSQL_PASSWORD"},
				Destination: &mysqlPassword,
			},
			&cli.StringFlag{
				Name:        "mysql-database",
				Value:       "event_service_test",
				EnvVars:     []string{"MYSQL_DATABASE"},
				Destination: &mysqlDatabase,
			},
			&cli.StringFlag{
				Name:        "mysql-protocol",
				Value:       "tcp",
				EnvVars:     []string{"MYSQL_PROTOCOL"},
				Destination: &mysqlProtocol,
			},
			&cli.StringFlag{
				Name:        "mysql-address",
				Value:       "127.0.0.1",
				EnvVars:     []string{"MYSQL_ADDRESS"},
				Destination: &mysqlAddress,
			},
			&cli.IntFlag{
				Name:        "mysql-max-open-conns",
				Value:       8,
				EnvVars:     []string{"MYSQL_MAX_OPEN_CONNS"},
				Destination: &mysqlMaxOpenConns,
			},
			&cli.IntFlag{
				Name:        "mysql-max-idle-conns",
				Value:       8,
				EnvVars:     []string{"MYSQL_MAX_IDLE_CONNS"},
				Destination: &mysqlMaxIdleConns,
			},
			&cli.IntFlag{
				Name:        "batch-size",
				Value:       500,
				EnvVars:     []string{"BATCH_SIZE"},
				Destination: &batchSize,
			},
			&cli.IntFlag{
				Name:        "batch-interval-milliseconds",
				Value:       10,
				EnvVars:     []string{"BATCH_INTERVAL_MILLISECONDS"},
				Destination: &batchIntervalMilliseconds,
			},
			&cli.Int64Flag{
				Name:        "starting-category-id",
				Value:       1,
				EnvVars:     []string{"STARTING_CATEGORY_ID"},
				Destination: &startingCategoryID,
			},
			&cli.IntFlag{
				Name:        "category-count",
				Value:       10000,
				EnvVars:     []string{"CATEGORY_COUNT"},
				Destination: &categoryCount,
			},
			&cli.TimestampFlag{
				Name:    "starting-category-create-time",
				Value:   cli.NewTimestamp(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
				EnvVars: []string{"STARTING_CATEGORY_CREATE_TIME"},
				Layout:  time.RFC3339,
			},
			&cli.TimestampFlag{
				Name:    "ending-category-create-time",
				Value:   cli.NewTimestamp(time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC)),
				EnvVars: []string{"ENDING_CATEGORY_CREATE_TIME"},
				Layout:  time.RFC3339,
			},
			&cli.Int64Flag{
				Name:        "starting-event-id",
				Value:       1,
				EnvVars:     []string{"starting_event_id"},
				Destination: &startingEventID,
			},
			&cli.IntFlag{
				Name:        "event-count",
				Value:       1000000,
				EnvVars:     []string{"EVENT_COUNT"},
				Destination: &eventCount,
			},
			&cli.TimestampFlag{
				Name:    "starting-event-create-time",
				Value:   cli.NewTimestamp(time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC)),
				EnvVars: []string{"STARTING_EVENT_CREATE_TIME"},
				Layout:  time.RFC3339,
			},
			&cli.TimestampFlag{
				Name:    "ending-event-create-time",
				Value:   cli.NewTimestamp(time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC)),
				EnvVars: []string{"ENDING_EVENT_CREATE_TIME"},
				Layout:  time.RFC3339,
			},
		},
		Action: func(clictx *cli.Context) error {
			mysqlConfig := mysql.NewConfig()
			mysqlConfig.User = mysqlUsername
			mysqlConfig.Passwd = mysqlPassword
			mysqlConfig.DBName = mysqlDatabase
			mysqlConfig.Net = mysqlProtocol
			mysqlConfig.Addr = mysqlAddress
			mysqlConfig.ParseTime = true

			sqlDB, err := sql.Open("mysql", mysqlConfig.FormatDSN())
			if err != nil {
				return fmt.Errorf("open mysql connection: %w", err)
			}

			sqlDB.SetMaxOpenConns(mysqlMaxOpenConns)
			sqlDB.SetMaxIdleConns(mysqlMaxIdleConns)

			startingCategoryCreateTime := clictx.Timestamp("starting-category-create-time")
			endingCategoryCreateTime := clictx.Timestamp("ending-category-create-time")
			startingEventCreateTime := clictx.Timestamp("starting-event-create-time")
			endingEventCreateTime := clictx.Timestamp("ending-event-create-time")

			initFakeData(
				sqlDB,
				batchSize,
				time.Duration(batchIntervalMilliseconds)*time.Millisecond,
				startingCategoryID,
				categoryCount,
				*startingCategoryCreateTime,
				*endingCategoryCreateTime,
				startingEventID,
				eventCount,
				*startingEventCreateTime,
				*endingEventCreateTime,
			)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}

func initFakeData(
	sqlDB *sql.DB,
	batchSize int,
	batchInterval time.Duration,
	startingCategoryID int64,
	categoryCount int,
	startingCategoryCreateTime time.Time,
	endingCategoryCreateTime time.Time,
	startingEventID int64,
	eventCount int,
	startingEventCreateTime time.Time,
	endingEventCreateTime time.Time,
) {
	fake := faker.New()

	if err := batch.All(categoryCount, batchSize, func(start, end int) error {
		if err := batchCreateRandomCategories(
			sqlDB,
			start,
			end,
			fake,
			startingCategoryID,
			startingCategoryCreateTime,
			endingCategoryCreateTime,
		); err != nil {
			log.Printf("Failed to batch create random categories: %v\n", err)
		}

		return nil
	}); err != nil {
		log.Printf("Failed to batch all: %v\n", err)
	}

	if err := batch.All(eventCount, batchSize, func(start, end int) error {
		if err := batchCreateRandomEvents(
			sqlDB,
			start,
			end,
			fake,
			startingEventID,
			startingEventCreateTime,
			endingEventCreateTime,
			startingCategoryID,
			categoryCount,
		); err != nil {
			log.Printf("Failed to batch create random events: %v\n", err)
		}

		return nil
	}); err != nil {
		log.Printf("Failed to batch all: %v\n", err)
	}
}

func batchCreateRandomEvents(
	sqlDB *sql.DB,
	start int,
	end int,
	fake faker.Faker,
	startingEventID int64,
	startingEventCreateTime time.Time,
	endingEventCreateTime time.Time,
	startingCategoryID int64,
	categoryCount int,
) error {
	tx, err := sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	for i := start; i <= end; i++ {
		if err := createRandomEvent(
			tx,
			fake,
			startingEventID+int64(i),
			startingEventCreateTime,
			endingEventCreateTime,
			startingCategoryID,
			categoryCount,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func createRandomEvent(
	tx db.DBTX,
	fake faker.Faker,
	eventID int64,
	startingEventCreateTime time.Time,
	endingEventCreateTime time.Time,
	startingCategoryID int64,
	categoryCount int,
) error {
	createTime := fake.Time().TimeBetween(startingEventCreateTime, endingEventCreateTime)
	startTime := fake.Time().TimeBetween(createTime, createTime.Add(7*24*time.Hour))

	var (
		endTime    time.Time
		updateTime time.Time
	)

	if fake.BoolWithChance(90) {
		endTime = fake.Time().TimeBetween(startTime, endTime.Add(24*time.Hour))
		updateTime = endTime
	} else {
		updateTime = createTime
	}

	categoryID := fake.Int64Between(startingCategoryID, int64(categoryCount)+startingCategoryID-1)

	_, err := db.
		New(tx).
		CreateEvent(
			context.TODO(),
			db.CreateEventParams{
				ID:        eventID,
				EnName:    fmt.Sprintf("Event %d", eventID),
				StartTime: startTime,
				EndTime: sql.NullTime{
					Time:  endTime,
					Valid: !endTime.IsZero(),
				},
				CategoryID: categoryID,
				CreateTime: createTime,
				UpdateTime: updateTime,
			},
		)
	if err != nil {
		return fmt.Errorf("create random event: %w", err)
	}

	return nil
}

func batchCreateRandomCategories(
	sqlDB *sql.DB,
	start int,
	end int,
	fake faker.Faker,
	startingCategoryID int64,
	startingCategoryCreateTime time.Time,
	endingCategoryCreateTime time.Time,
) error {
	tx, err := sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	for i := start; i <= end; i++ {
		if err := createRandomCategory(
			tx,
			fake,
			startingCategoryID+int64(i),
			startingCategoryCreateTime,
			endingCategoryCreateTime,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func createRandomCategory(
	tx db.DBTX,
	fake faker.Faker,
	categoryID int64,
	startingCategoryCreateTime time.Time,
	endingCategoryCreateTime time.Time,
) error {
	createTime := fake.Time().TimeBetween(startingCategoryCreateTime, endingCategoryCreateTime)
	updateTime := fake.Time().TimeBetween(createTime, createTime.Add(24*time.Hour))

	_, err := db.
		New(tx).
		CreateCategory(
			context.TODO(),
			db.CreateCategoryParams{
				ID:         categoryID,
				EnName:     fmt.Sprintf("Category %d", categoryID),
				CreateTime: createTime,
				UpdateTime: updateTime,
			},
		)
	if err != nil {
		return fmt.Errorf("create random category: %w", err)
	}

	return nil
}
