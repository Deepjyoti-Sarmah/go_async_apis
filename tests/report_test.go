package tests

import (
	"context"
	"testing"
	"time"

	"github.com/Deepjyoti-Sarmah/fast-api/fixtures"
	"github.com/Deepjyoti-Sarmah/fast-api/store"
	"github.com/stretchr/testify/require"
)

func TestReportStore(t *testing.T) {
	env := fixtures.NewTestEnv(t)
	cleanup := env.SetupDb(t)
	t.Cleanup(func() {
		cleanup(t)
	})

	ctx := context.Background()
	reportStore := store.NewReportStore(env.Db)
	userStore := store.NewUserStore(env.Db)
	user, err := userStore.CreateUser(ctx, "testing@test.com", "secretpassword")
	require.NoError(t, err)

	now := time.Now()
	report, err := reportStore.Create(ctx, user.Id, "monsters")
	require.NoError(t, err)
	require.Equal(t, user.Id, report.UserId)
	require.Equal(t, "monsters", report.ReportType)
	require.Less(t, now.UnixNano(), report.CreatedAt.UnixNano())

	startedAt := report.CreatedAt.Add(time.Second)
	completedAt := report.CreatedAt.Add(2 * time.Second)
	failedAt := report.CreatedAt.Add(3 * time.Second)
	errorMsg := "there was a failure"
	downloadUrl := "http://localhost:8080/reports"
	outputPath := "s3://reeports-test/reports"
	downloadUrlExpiresAt := report.CreatedAt.Add(4 * time.Second)

	report.ReportType = "food"
	report.StartedAt = &startedAt
	report.CompletedAt = &completedAt
	report.FailedAt = &failedAt
	report.ErrorMessage = &errorMsg
	report.DownloadUrl = &downloadUrl
	report.OutputFilePath = &outputPath
	report.DownloadUrlExpiresAt = &downloadUrlExpiresAt

	report2, err := reportStore.Update(ctx, report)
	require.NoError(t, err)

	require.Equal(t, report.UserId, report2.UserId)
	require.Equal(t, report.Id, report2.Id)
	require.Equal(t, "monsters", report2.ReportType)
	require.Equal(t, report.CreatedAt.UnixNano(), report2.CreatedAt.UnixNano())
	require.Equal(t, report.StartedAt.UnixNano(), report2.StartedAt.UnixNano())
	require.Equal(t, report.CompletedAt.UnixNano(), report2.CompletedAt.UnixNano())
	require.Equal(t, report.FailedAt.UnixNano(), report2.FailedAt.UnixNano())
	require.Equal(t, report.ErrorMessage, &errorMsg)
	require.Equal(t, report.ErrorMessage, &errorMsg)
	require.Equal(t, report.DownloadUrl, &downloadUrl)
	require.Equal(t, report.OutputFilePath, &outputPath)
	require.Equal(t, report.DownloadUrlExpiresAt.UnixNano(), (&downloadUrlExpiresAt).UnixNano())

	report3, err := reportStore.ByPrimaryKey(ctx, report.UserId, report.Id)
	require.NoError(t, err)

	require.Equal(t, report.UserId, report3.UserId)
	require.Equal(t, report.Id, report3.Id)
	require.Equal(t, "monsters", report3.ReportType)
	require.Equal(t, report.CreatedAt.UnixNano(), report3.CreatedAt.UnixNano())
	require.Equal(t, report.StartedAt.UnixNano(), report3.StartedAt.UnixNano())
	require.Equal(t, report.CompletedAt.UnixNano(), report3.CompletedAt.UnixNano())
	require.Equal(t, report.FailedAt.UnixNano(), report3.FailedAt.UnixNano())
	require.Equal(t, report3.ErrorMessage, &errorMsg)
	require.Equal(t, report3.ErrorMessage, &errorMsg)
	require.Equal(t, report3.DownloadUrl, &downloadUrl)
	require.Equal(t, report3.OutputFilePath, &outputPath)
	require.Equal(t, report3.DownloadUrlExpiresAt.UnixNano(), (&downloadUrlExpiresAt).UnixNano())
}
