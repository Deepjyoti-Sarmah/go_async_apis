package tests

import (
	"context"
	"testing"

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

	report, err := reportStore.Create(ctx, user.Id, "monsters")
	require.NoError(t, err)
	require.Equal(t, user.Id, report.UserId)
	require.Equal(t, "monsters", report.ReportType)
}
