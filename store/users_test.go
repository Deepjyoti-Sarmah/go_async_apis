package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Deepjyoti-Sarmah/fast-api/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T) {
	os.Setenv("ENV", string(config.Env_Test))
	conf, err := config.New()
	require.NoError(t, err)

	db, err := NewPostgresDb(conf)
	require.NoError(t, err)
	defer db.Close()

	currentDir, err := os.Getwd()
	require.NoError(t, err)

	projectRoot := filepath.Dir(currentDir)
	migrationsPath := filepath.Join(projectRoot, "migrations")

	_, err = os.Stat(migrationsPath)
	require.NoError(t, err, "migrations directory not found at: "+migrationsPath)

	m, err := migrate.New(
		// "file://./migrations",
		"file://"+migrationsPath,
		conf.DatabaseUrl(),
	)
	require.NoError(t, err)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

	userStore := NewUserStore(db)
	user, err := userStore.CreateUser(context.Background(), "test@test.com", "testingpassword")
	require.NoError(t, err)

	require.Equal(t, "test@test.com", user.Email)
	require.NoError(t, user.ComparePassword("testingpassword"))
}
