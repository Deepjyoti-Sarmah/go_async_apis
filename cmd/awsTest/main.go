package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Deepjyoti-Sarmah/fast-api/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
}
