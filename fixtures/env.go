package fixtures

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Deepjyoti-Sarmah/fast-api/config"
	"github.com/Deepjyoti-Sarmah/fast-api/store"
	"github.com/golang-migrate/migrate/v4"
	"github.com/stretchr/testify/require"
)

type TestEnv struct {
	Config *config.Config
	Db     *sql.DB
}

func NewTestEnv(t *testing.T) *TestEnv {
	os.Setenv("ENV", string(config.Env_Test))
	conf, err := config.New()
	require.NoError(t, err)

	db, err := store.NewPostgresDb(conf)
	require.NoError(t, err)

	return &TestEnv{
		conf,
		db,
	}
}

func (te *TestEnv) SetupDb(t *testing.T) func(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	projectRoot := filepath.Dir(currentDir)
	migrationsPath := filepath.Join(projectRoot, "migrations")

	_, err = os.Stat(migrationsPath)
	require.NoError(t, err, "migrations directory not found at: "+migrationsPath)

	m, err := migrate.New(
		// "file://./migrations",
		"file://"+migrationsPath,
		te.Config.DatabaseUrl(),
	)
	require.NoError(t, err)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

	return te.TeardownDb
}

func (te *TestEnv) TeardownDb(t *testing.T) {
	_, err := te.Db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", strings.Join([]string{"users", "refresh_tokens", "reports"}, ", ")))
	require.NoError(t, err)
}
