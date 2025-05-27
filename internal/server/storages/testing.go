package storages

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
)

func RunTestPgContainer(t testing.TB, ctx context.Context, cfg *config.Config) *postgres.PostgresContainer {
	t.Helper()
	pgconf, err := pgx.ParseConfig(cfg.DatabaseDSN)
	if err != nil {
		t.Errorf("failed to parse database dsn %s", err.Error())
	}

	pgContainer, err := postgres.Run(ctx,
		"postgres:15.3-alpine",
		postgres.WithDatabase(pgconf.Database),
		postgres.WithUsername(pgconf.User),
		postgres.WithPassword(pgconf.Password),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Errorf("failed to run pg container error %s", err.Error())
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})

	return pgContainer
}
