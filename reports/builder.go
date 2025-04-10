package reports

import (
	"github.com/Deepjyoti-Sarmah/fast-api/store"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type RepportBuilder struct {
	reportStore *store.ReportStore
	LozClient   *LozClient
	s3Client    *s3.Client
}

func NewReportBuilder(reportStore *store.ReportStore, lozClient *LozClient, s3Client *s3.Client) *RepportBuilder {
	return &RepportBuilder{
		reportStore: reportStore,
		LozClient:   lozClient,
		s3Client:    s3Client,
	}
}
