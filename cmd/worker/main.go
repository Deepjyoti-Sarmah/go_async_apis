package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Deepjyoti-Sarmah/fast-api/config"
	"github.com/Deepjyoti-Sarmah/fast-api/reports"
	"github.com/Deepjyoti-Sarmah/fast-api/store"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := config.New()
	if err != nil {
		return err
	}

	db, err := store.NewPostgresDb(config)
	if err != nil {
		return err
	}
	dataStore := store.New(db)

	awsConf, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	s3Client := s3.NewFromConfig(awsConf, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(config.S3LocalstackEndpont)
		options.UsePathStyle = true
	})

	sqsClient := sqs.NewFromConfig(awsConf, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(config.LocalstackEndpont)
	})

	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(jsonHandler)

	lozClient := reports.NewLozClient(&http.Client{Timeout: time.Second * 10})
	builder := reports.NewReportBuilder(config, dataStore.ReportStore, lozClient, s3Client, logger)

	maxConcurrency := 2
	worker := reports.NewWorker(config, logger, builder, sqsClient, maxConcurrency)

	if err := worker.Start(ctx); err != nil {
		return err
	}

	return nil
}
