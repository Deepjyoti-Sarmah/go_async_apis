package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Deepjyoti-Sarmah/fast-api/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Worker struct {
	config      *config.Config
	builder     *ReportBuilder
	logger      *slog.Logger
	sqsClient   *sqs.Client
	channel     chan types.Message
	concurrency int
}

func NewWorker(config *config.Config, logger *slog.Logger, builder *ReportBuilder, sqsClient *sqs.Client, maxConcurrency int) *Worker {
	return &Worker{
		config:    config,
		logger:    logger,
		builder:   builder,
		sqsClient: sqsClient,
		channel:   make(chan types.Message, maxConcurrency),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	queueUrlOutput, err := w.sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(w.config.SqsQueue),
	})
	if err != nil {
		return fmt.Errorf("failed to get url for queue %s: %w", w.config.SqsQueue, err)
	}

	w.logger.Info("started Worker", "queue", w.config.SqsQueue, "queue_url", queueUrlOutput.QueueUrl)

	// consummer
	for i := 0; i < w.concurrency; i++ {
		go func(id int) {
			for {
				select {
				case <-ctx.Done():
					w.logger.Error("worker stopped", "goroutine_id", id, "error", ctx.Err())
					return
				case message := <-w.channel:
					if err := w.processMessage(ctx, message, queueUrlOutput.QueueUrl); err != nil {
						w.logger.Error("failed to process message", "error", err, "goroutine_id", id)
					}
				}
			}
		}(i)
	}

	// producer
	for {
		output, err := w.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            queueUrlOutput.QueueUrl,
			MaxNumberOfMessages: int32(w.concurrency + 1),
		})
		if err != nil {
			w.logger.Error("failed to receive messgae", "error", err)
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}

		if len(output.Messages) == 0 {
			continue
		}

		for _, message := range output.Messages {
			w.channel <- message
		}
	}
}

func (w *Worker) processMessage(ctx context.Context, message types.Message, queueUrl *string) error {
	w.logger.Info("processing message", "message_id", *message.MessageId)
	if message.Body == nil || *message.Body == "" {
		w.logger.Warn("message body is invalid", "messgae_id", message.MessageId)
		return nil
	}

	var msg SqsMessage
	if err := json.Unmarshal([]byte(*message.Body), &msg); err != nil {
		w.logger.Warn("message body is invalid", "messgae_id", message.MessageId, "body", *message.Body)
		return nil
	}

	builderCtx, builderCancel := context.WithTimeout(ctx, time.Second*10)
	defer builderCancel()
	_, err := w.builder.Build(builderCtx, msg.UserId, msg.ReportId)
	if err != nil {
		return fmt.Errorf("failed to build report: %w", err)
	}

	if _, err := w.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      queueUrl,
		ReceiptHandle: message.ReceiptHandle,
	}); err != nil {
		return fmt.Errorf("failed to delete message %s %w", *message.MessageId, err)
	}

	return nil
}
