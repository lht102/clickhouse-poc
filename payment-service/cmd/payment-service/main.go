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
	"github.com/lht102/clickhouse-poc/payment-service/db"
	"github.com/pacedotdev/batch"
	"github.com/urfave/cli/v2"
)

type PaymentStatus string

type TransactionStatus string

const (
	PaymentStatusAccepted  PaymentStatus = "ACCEPTED"
	PaymentStatusRejected  PaymentStatus = "REJECTED"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"

	TransactionStatusPending TransactionStatus = "PENDING"
	TransactionStatusSettled TransactionStatus = "SETTLED"
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

		batchSize                     int
		batchIntervalMilliseconds     int
		startingPaymentID             int64
		paymentCount                  int
		minTransactionCountPerPayment int
		maxTransactionCountPerPayment int
		startingEventID               int64
		endingEventID                 int64
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
				Name:        "starting-payment-id",
				Value:       1,
				EnvVars:     []string{"STARTING_PAYMENT_ID"},
				Destination: &startingPaymentID,
			},
			&cli.IntFlag{
				Name:        "payment-count",
				Value:       10000000,
				EnvVars:     []string{"PAYMENT_COUNT"},
				Destination: &paymentCount,
			},
			&cli.TimestampFlag{
				Name:    "starting-payment-create-time",
				Value:   cli.NewTimestamp(time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC)),
				EnvVars: []string{"STARTING_PAYMENT_CREATE_TIME"},
				Layout:  time.RFC3339,
			},
			&cli.TimestampFlag{
				Name:    "ending-payment-create-time",
				Value:   cli.NewTimestamp(time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)),
				EnvVars: []string{"ENDING_PAYMENT_CREATE_TIME"},
				Layout:  time.RFC3339,
			},
			&cli.IntFlag{
				Name:        "min-transaction-count-per-payment",
				Value:       1,
				EnvVars:     []string{"MIN_TRANSACTION_COUNT_PER_PAYMENT"},
				Destination: &minTransactionCountPerPayment,
			},
			&cli.IntFlag{
				Name:        "max-transaction-count-per-payment",
				Value:       3,
				EnvVars:     []string{"MAX_TRANSACTION_COUNT_PER_PAYMENT"},
				Destination: &maxTransactionCountPerPayment,
			},
			&cli.Int64Flag{
				Name:        "starting-event-id",
				Value:       1,
				EnvVars:     []string{"STARTING_EVENT_ID"},
				Destination: &startingEventID,
			},
			&cli.Int64Flag{
				Name:        "ending-event-id",
				Value:       1000000,
				EnvVars:     []string{"ENDING_EVENT_ID"},
				Destination: &endingEventID,
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

			startingPaymentCreateTime := clictx.Timestamp("starting-payment-create-time")
			endingPaymentCreateTime := clictx.Timestamp("ending-payment-create-time")

			initFakeData(
				sqlDB,
				batchSize,
				time.Duration(batchIntervalMilliseconds)*time.Millisecond,
				startingPaymentID,
				paymentCount,
				*startingPaymentCreateTime,
				*endingPaymentCreateTime,
				minTransactionCountPerPayment,
				maxTransactionCountPerPayment,
				startingEventID,
				endingEventID,
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
	startingPaymentID int64,
	paymentCount int,
	startingPaymentCreateTime time.Time,
	endingPaymentCreateTime time.Time,
	minTransactionCountPerPayment int,
	maxTransactionCountPerPayment int,
	startingEventID int64,
	endingEventID int64,
) {
	fake := faker.New()

	if err := batch.All(paymentCount, batchSize, func(start, end int) error {
		if err := batchCreateRandomPayments(
			sqlDB,
			start,
			end,
			fake,
			startingPaymentID,
			startingPaymentCreateTime,
			endingPaymentCreateTime,
			startingEventID,
			endingEventID,
			minTransactionCountPerPayment,
			maxTransactionCountPerPayment,
		); err != nil {
			log.Printf("Failed to batch create random payments: %v\n", err)
		}

		return nil
	}); err != nil {
		log.Printf("Failed to batch all: %v\n", err)
	}
}

func batchCreateRandomPayments(
	sqlDB *sql.DB,
	start int,
	end int,
	fake faker.Faker,
	startingPaymentID int64,
	startingPaymentCreateTime time.Time,
	endingPaymentCreateTime time.Time,
	startingEventID int64,
	endingEventID int64,
	minTransactionCountPerPayment int,
	maxTransactionCountPerPayment int,
) error {
	tx, err := sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	for i := start; i <= end; i++ {
		if err := createRandomPayment(
			tx,
			fake,
			startingPaymentID+int64(i),
			startingPaymentCreateTime,
			endingPaymentCreateTime,
			startingEventID,
			endingEventID,
			minTransactionCountPerPayment,
			maxTransactionCountPerPayment,
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

func createRandomPayment(
	tx db.DBTX,
	fake faker.Faker,
	paymentID int64,
	startingPaymentCreateTime time.Time,
	endingPaymentCreateTime time.Time,
	startingEventID int64,
	endingEventID int64,
	minTransactionCountPerPayment int,
	maxTransactionCountPerPayment int,
) error {
	paymentStatus := generateRandomPaymentStatus(fake)
	paymentCreateTime := fake.Time().TimeBetween(startingPaymentCreateTime, endingPaymentCreateTime)

	var (
		paymentUpdateTime   time.Time
		paymentCompleteTime time.Time
	)

	if paymentStatus == PaymentStatusCompleted {
		paymentCompleteTime = fake.Time().TimeBetween(paymentCreateTime, paymentCreateTime.Add(7*24*time.Hour))
		paymentUpdateTime = paymentCompleteTime
	} else {
		paymentUpdateTime = fake.Time().TimeBetween(paymentCreateTime, paymentCreateTime.Add(3*24*time.Hour))
	}

	_, err := db.
		New(tx).
		CreatePayment(
			context.TODO(),
			db.CreatePaymentParams{
				ID:            paymentID,
				PaymentStatus: string(paymentStatus),
				CompleteTime: sql.NullTime{
					Time:  paymentCompleteTime,
					Valid: !paymentCompleteTime.IsZero(),
				},
				CreateTime: paymentCreateTime,
				UpdateTime: paymentUpdateTime,
			},
		)
	if err != nil {
		return fmt.Errorf("create random payment: %w", err)
	}

	for _, params := range generateRandomTransactionParamsList(
		fake,
		paymentID,
		startingEventID,
		endingEventID,
		minTransactionCountPerPayment,
		maxTransactionCountPerPayment,
		paymentStatus,
		paymentCreateTime,
		paymentCompleteTime,
	) {
		_, err := db.
			New(tx).
			CreateTransaction(
				context.TODO(),
				params,
			)
		if err != nil {
			return fmt.Errorf("create random transaction: %w", err)
		}
	}

	return nil
}

func generateRandomTransactionParamsList(
	fake faker.Faker,
	paymentID int64,
	startingEventID int64,
	endingEventID int64,
	minTransactionCountPerPayment int,
	maxTransactionCountPerPayment int,
	paymentStatus PaymentStatus,
	paymentCreateTime time.Time,
	paymentCompleteTime time.Time,

) []db.CreateTransactionParams {
	transactionCount := fake.IntBetween(minTransactionCountPerPayment, maxTransactionCountPerPayment)
	if transactionCount == 0 {
		return nil
	}

	list := make([]db.CreateTransactionParams, 0, transactionCount)

	transactionStatuses := make([]TransactionStatus, 0, transactionCount)
	for i := 0; i < transactionCount; i++ {
		transactionStatuses = append(transactionStatuses, generateRandomTransactionStatus(fake))
	}

	switch paymentStatus {
	case PaymentStatusAccepted, PaymentStatusRejected:
		transactionStatuses[fake.IntBetween(0, transactionCount-1)] = TransactionStatusPending
	case PaymentStatusCompleted:
		for i := 0; i < transactionCount; i++ {
			transactionStatuses[i] = TransactionStatusSettled
		}
	}

	for i := 0; i < transactionCount; i++ {
		transactionCreateTime := fake.Time().TimeBetween(
			paymentCreateTime,
			maxTime(paymentCreateTime.Add(time.Hour), paymentCompleteTime),
		)
		transactionUpdateTime := paymentCompleteTime
		transactionUpdateTime = fake.Time().TimeBetween(
			transactionCreateTime,
			minTime(transactionCreateTime.Add(3*24*time.Hour), transactionUpdateTime),
		)

		list = append(list, db.CreateTransactionParams{
			TransactionStatus: string(transactionStatuses[i]),
			EventID:           fake.Int64Between(startingEventID, endingEventID),
			PaymentID:         paymentID,
			CreateTime:        transactionCreateTime,
			UpdateTime:        transactionUpdateTime,
		})
	}

	return list
}

func generateRandomPaymentStatus(fake faker.Faker) PaymentStatus {
	return PaymentStatus(fake.RandomStringElement(
		[]string{
			string(PaymentStatusAccepted),
			string(PaymentStatusRejected),
			string(PaymentStatusCompleted),
		},
	))
}

func generateRandomTransactionStatus(fake faker.Faker) TransactionStatus {
	return TransactionStatus(fake.RandomStringElement(
		[]string{
			string(TransactionStatusPending),
			string(TransactionStatusSettled),
		},
	))
}

func maxTime(x time.Time, y time.Time) time.Time {
	if x.After(y) {
		return x
	}

	return y
}

func minTime(x time.Time, y time.Time) time.Time {
	if x.After(y) {
		return y
	}

	return x
}
