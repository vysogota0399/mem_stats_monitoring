package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"golang.org/x/sync/errgroup"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage/interfaces"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

// All возвращает все записи из базы данных.
func (s *DBStorage) All() map[string]map[string][]string {
	s.lg.WarnCtx(context.Background(), "useless method")
	return make(map[string]map[string][]string)
}

// Last возвращает последнюю запись из базы данных.
// Deprecated: исторически так сложилось.
func (s *DBStorage) Last(mType, mName string) (string, error) {
	s.lg.WarnCtx(context.Background(), "useless method")
	return "", nil
}

// Push добавляет новую заись в базу данных.
// Deprecated: исторически так сложилось.
func (s *DBStorage) Push(mType, mName string, val any) error {
	s.lg.WarnCtx(context.Background(), "useless method")
	return nil
}

// Ping проверяет соединение с бащой данных.
func (s *DBStorage) Ping() error {
	return s.db.Ping()
}

type Migrator interface {
	Migrate(db interfaces.IDB) error
}

type GooseMigrator struct {
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func (m *GooseMigrator) Migrate(db interfaces.IDB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return err
	}

	sqldb, ok := db.(*sql.DB)
	if !ok {
		return fmt.Errorf("db: db is not *sql.DB")
	}

	if err := goose.Up(sqldb, "migrations"); err != nil {
		return err
	}

	return nil
}

type ConnectionOpener interface {
	OpenDB(ctx context.Context, dbDsn string) (interfaces.IDB, error)
}

type PGConnectionOpener struct {
	atpt           uint8
	maxOpenRetries uint8
	lg             *logging.ZapLogger
	dbDsn          string
}

func (o *PGConnectionOpener) OpenDB(ctx context.Context, dbDsn string) (interfaces.IDB, error) {
	o.atpt++
	db, err := sql.Open(pgxDriver, dbDsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			o.lg.ErrorCtx(context.Background(), "failed to close db after ping error", zap.Error(closeErr))
		}
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) && o.atpt <= o.maxOpenRetries {
			time.Sleep(time.Duration(utils.Delay(o.atpt)) * time.Second)
			return o.OpenDB(ctx, o.dbDsn)
		}

		return nil, err
	}

	return db, nil
}

const pgxDriver string = "pgx"

// DBStorage реализация интерфейса Storage. В качестве зранилища используется база данных postgresql.
type DBStorage struct {
	dbDsn            string // строка подключения к базе данных.
	db               interfaces.IDB
	lg               *logging.ZapLogger
	maxOpenRetries   uint8 // максимальное количетсво попыток открыть соединение.
	migrator         Migrator
	connectionOpener ConnectionOpener
	testMode         bool
}

func NewDBStorage(
	ctx context.Context,
	cfg config.Config,
	errg *errgroup.Group,
	lg *logging.ZapLogger,
	migrator Migrator,
	connectionOpener ConnectionOpener,
) (Storage, error) {
	strg := &DBStorage{
		lg:               lg,
		dbDsn:            cfg.DatabaseDSN,
		maxOpenRetries:   4,
		migrator:         migrator,
		connectionOpener: connectionOpener,
	}

	db, err := strg.connectionOpener.OpenDB(context.Background(), strg.dbDsn)
	if err != nil {
		return nil, fmt.Errorf("db: open db failed error %w", err)
	}

	strg.db = db

	if err := strg.migrator.Migrate(db); err != nil {
		if closeErr := strg.db.Close(); closeErr != nil {
			lg.ErrorCtx(ctx, "failed to close db after migration error", zap.Error(closeErr))
		}
		return nil, err
	}

	ctx = lg.WithContextFields(ctx, zap.String("name", "db_storage"))

	errg.Go(func() error {
		<-ctx.Done()
		if err := strg.db.Close(); err != nil {
			return fmt.Errorf("db: close db failed error %w", err)
		}

		return nil
	})

	return strg, nil
}

type ResultFunc func(rows *sql.Rows) error

func (s *DBStorage) QueryRowContext(ctx context.Context, query string, args []any, result ResultFunc) error {
	rows, err := s.db.QueryContext(ctx, query, args)
	if err != nil {
		return err
	}
	defer func() {
		if s.testMode {
			return
		}
		if closeErr := rows.Close(); closeErr != nil {
			s.lg.ErrorCtx(ctx, "failed to close rows", zap.Error(closeErr))
		}
	}()

	if err := result(rows); err != nil {
		return fmt.Errorf("db: query row result error %w", err)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("db: query row error %w", err)
	}

	return nil
}

func (s *DBStorage) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, opts)
}

func (s *DBStorage) CommitTx(ctx context.Context, tx *sql.Tx) error {
	return tx.Commit()
}

func (s *DBStorage) RollbackTx(ctx context.Context, tx *sql.Tx) error {
	return tx.Rollback()
}
