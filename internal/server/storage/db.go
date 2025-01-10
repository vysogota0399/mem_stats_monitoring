package storage

import (
	"context"
	"database/sql"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type DBStorage struct {
	db *sql.DB
	lg *logging.ZapLogger
}

func (s *DBStorage) All() map[string]map[string][]string {
	s.lg.WarnCtx(context.Background(), "useless method")
	return make(map[string]map[string][]string)
}

func (s *DBStorage) Last(mType, mName string) (string, error) {
	return "", nil
}

func (s *DBStorage) Push(mType, mName string, val any) error {
	return nil
}

func (s *DBStorage) Ping() error {
	return s.db.Ping()
}

func (s *DBStorage) DB() *sql.DB {
	return s.db
}

type DBAble interface {
	Storage
	Ping() error
	DB() *sql.DB
}

const pgxDriver string = "pgx"

func NewDBStorage(ctx context.Context, cfg config.Config, wg *sync.WaitGroup, lg *logging.ZapLogger) (Storage, error) {
	db, err := sql.Open(pgxDriver, cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	strg := &DBStorage{
		db: db,
		lg: lg,
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	ctx = lg.WithContextFields(ctx, zap.String("name", "db_storage"))
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		if err := db.Close(); err != nil {
			lg.FatalCtx(
				ctx,
				"close db failed",
				zap.Error(err),
			)
		}
	}()

	return strg, nil
}
