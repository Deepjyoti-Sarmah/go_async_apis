package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Deepjyoti-Sarmah/fast-api/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	ctx := context.Background()
	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("Couldn't land default configuration. Have you set up your aws account?")
		fmt.Println(err)
		return
	}

	conf, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(sdkConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(conf.S3LocalstackEndpont)
		options.UsePathStyle = true
	})
	out, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal("err")
	}

	for _, bucket := range out.Buckets {
		fmt.Println(*bucket.Name)
	}

	sqsClient := sqs.NewFromConfig(sdkConfig, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(conf.LocalstackEndpont)
	})

	sqsOut, err := sqsClient.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, q := range sqsOut.QueueUrls {
		fmt.Println(q)
	}
}
