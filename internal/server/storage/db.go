package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type DBStorage struct {
	dbDsn          string
	db             *sql.DB
	lg             *logging.ZapLogger
	maxOpenRetries uint8
}

func (s *DBStorage) All() map[string]map[string][]string {
	s.lg.WarnCtx(context.Background(), "useless method")
	return make(map[string]map[string][]string)
}

func (s *DBStorage) Last(mType, mName string) (string, error) {
	s.lg.WarnCtx(context.Background(), "useless method")
	return "", nil
}

func (s *DBStorage) Push(mType, mName string, val any) error {
	s.lg.WarnCtx(context.Background(), "useless method")
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
	strg := &DBStorage{
		lg:             lg,
		dbDsn:          cfg.DatabaseDSN,
		maxOpenRetries: 4,
	}

	if err := strg.openDB(0); err != nil {
		return nil, err
	}

	if err := strg.migrate(); err != nil {
		strg.db.Close()
		return nil, err
	}

	ctx = lg.WithContextFields(ctx, zap.String("name", "db_storage"))
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-ctx.Done()
		if err := strg.db.Close(); err != nil {
			lg.FatalCtx(
				ctx,
				"close db failed",
				zap.Error(err),
			)
		}
	}()

	return strg, nil
}

func (s *DBStorage) openDB(atpt uint8) error {
	atpt++
	db, err := sql.Open(pgxDriver, s.dbDsn)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) && atpt <= s.maxOpenRetries {
			time.Sleep(time.Duration(utils.Delay(atpt)) * time.Second)
			return s.openDB(atpt)
		}

		return err
	}

	s.db = db
	return nil
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func (s *DBStorage) migrate() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return err
	}

	if err := goose.Up(s.db, "migrations"); err != nil {
		return err
	}

	return nil
}
