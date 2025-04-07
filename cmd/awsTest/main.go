package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("Couldn't land default configuration. Have you set up your aws account?")
		fmt.Println(err)
		return
	}

	s3Client := s3.NewFromConfig(sdkConfig)
}
