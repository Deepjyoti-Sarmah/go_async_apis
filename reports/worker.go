package reports

import (
	"log/slog"

	"github.com/Deepjyoti-Sarmah/fast-api/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Worker struct {
	config    *config.Config
	builder   *ReportBuilder
	logger    *slog.Logger
	sqsClient *sqs.Client
	channel   chan types.Message
}

func NewWorker(config *config.Config, logger *slog.Logger, sqsClient *sqs.Client, maxComcurrency int) *Worker {
	return &Worker{
		config:    config,
		logger:    logger,
		sqsClient: sqsClient,
		channel:   make(chan types.Message, maxComcurrency),
	}
}


