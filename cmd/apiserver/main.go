package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Deepjyoti-Sarmah/fast-api/apiserver"
	"github.com/Deepjyoti-Sarmah/fast-api/config"
	"github.com/Deepjyoti-Sarmah/fast-api/store"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conf, err := config.New()
	if err != nil {
		return err
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(jsonHandler)

	db, err := store.NewPostgresDb(conf)
	if err != nil {
		return err
	}

	dataStore := store.New(db)
	jwtManager := apiserver.NewJwtManager(conf)

	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("couldn't load default configuration: %w", err)
	}

	sqsClient := sqs.NewFromConfig(sdkConfig, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(conf.LocalstackEndpont)
	})

	server := apiserver.New(conf, logger, dataStore, jwtManager, sqsClient)
	if err := server.Start(ctx); err != nil {
		return err
	}

	return nil
}
