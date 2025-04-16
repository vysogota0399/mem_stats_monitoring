package storages

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
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Migrator interface {
	Migrate(db *sql.DB) error
}

type GooseMigrator struct {
}

func NewGooseMigrator() *GooseMigrator {
	return &GooseMigrator{}
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func (m *GooseMigrator) Migrate(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return err
	}

	sqldb := db

	if err := goose.Up(sqldb, "migrations"); err != nil {
		return err
	}

	return nil
}

type ConnectionOpener interface {
	OpenDB(ctx context.Context, dbDsn string) (*sql.DB, error)
}

type PGConnectionOpener struct {
	atpt           uint8
	maxOpenRetries uint8
	lg             *logging.ZapLogger
	dbDsn          string
}

func NewPGConnectionOpener(lg *logging.ZapLogger, cfg *config.Config) *PGConnectionOpener {
	return &PGConnectionOpener{
		atpt:           10,
		maxOpenRetries: 10,
		lg:             lg,
		dbDsn:          cfg.DatabaseDSN,
	}
}

const pgxDriver string = "pgx"

func (o *PGConnectionOpener) OpenDB(ctx context.Context, dbDsn string) (*sql.DB, error) {
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

var _ ConnectionOpener = (*PGConnectionOpener)(nil)

type PG struct {
	dbDsn            string // строка подключения к базе данных.
	db               *sql.DB
	lg               *logging.ZapLogger
	maxOpenRetries   uint8 // максимальное количетсво попыток открыть соединение.
	migrator         Migrator
	connectionOpener ConnectionOpener
}

func NewPG(
	lc fx.Lifecycle,
	lg *logging.ZapLogger,
	cfg *config.Config,
	migrator Migrator,
	connectionOpener ConnectionOpener,
) (*PG, error) {
	db, err := connectionOpener.OpenDB(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	strg := &PG{
		db:               db,
		lg:               lg,
		dbDsn:            cfg.DatabaseDSN,
		maxOpenRetries:   5,
		migrator:         migrator,
		connectionOpener: connectionOpener,
	}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				if err := db.Ping(); err != nil {
					if closeErr := db.Close(); closeErr != nil {
						return errors.Join(fmt.Errorf("pg: close db error %w", closeErr), err)
					}

					return err
				}

				if err := migrator.Migrate(db); err != nil {
					return fmt.Errorf("pg: migrate failed error %w", err)
				}
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return strg.db.Close()
			},
		},
	)

	return strg, nil
}

func (pg *PG) CreateOrUpdate(ctx context.Context, mType, mName string, val any) error {
	var table string
	switch mType {
	case models.GaugeType:
		table = "gauges"
	case models.CounterType:
		table = "counters"
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET value = $2
	`, table)

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, err := pg.db.ExecContext(dbCtx, query, mName, val)
	if err != nil {
		return fmt.Errorf("pg: create or update failed error %w", err)
	}

	return nil
}

func (pg *PG) GetGauge(ctx context.Context, record *models.Gauge) error {
	query := `
		SELECT value, id, name
		FROM gauges
		WHERE name = $1
		ORDER BY id DESC
		LIMIT 1`

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rows, err := pg.db.QueryContext(dbCtx, query, record.Name)
	if err != nil {
		return fmt.Errorf("pg: get gauge failed error %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			pg.lg.ErrorCtx(ctx, "failed to close rows", zap.Error(closeErr))
		}
	}()

	if err := rows.Scan(&record.Value, &record.ID, &record.Name); err != nil {
		return fmt.Errorf("pg: get gauge failed error %w", err)
	}

	return nil
}

func (pg *PG) GetCounter(ctx context.Context, record *models.Counter) error {
	query := `
		SELECT value, id, name
		FROM counters
		WHERE name = $1
		ORDER BY id DESC
		LIMIT 1`

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rows, err := pg.db.QueryContext(dbCtx, query, record.Name)
	if err != nil {
		return fmt.Errorf("pg: get counter failed error %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			pg.lg.ErrorCtx(ctx, "failed to close rows", zap.Error(closeErr))
		}
	}()

	for rows.Next() {
		if err := rows.Scan(&record.Value, &record.ID, &record.Name); err != nil {
			return fmt.Errorf("pg: get counter failed error %w", err)
		}
	}

	return nil
}

func (pg *PG) GetGauges(ctx context.Context) ([]models.Gauge, error) {
	gauges := make([]models.Gauge, 0)

	rows, err := pg.db.QueryContext(ctx, `
		SELECT id, name, value
		FROM gauges
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("pg: get all gauges failed error %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			pg.lg.ErrorCtx(ctx, "db: failed to close rows", zap.Error(closeErr))
		}
	}()

	for rows.Next() {
		g := models.Gauge{}

		if err := rows.Scan(&g.ID, &g.Name, &g.Value); err != nil {
			return nil, fmt.Errorf("pg: get all gauges failed error %w", err)
		}
		gauges = append(gauges, g)
	}
	return gauges, nil
}

func (pg *PG) GetCounters(ctx context.Context) ([]models.Counter, error) {
	counters := make([]models.Counter, 0)

	rows, err := pg.db.QueryContext(ctx, `
		SELECT id, name, value
		FROM counters
		ORDER BY id DESC
	`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(ErrNoRecords, fmt.Errorf("pg: get all counters failed error %w", err))
		}

		return nil, fmt.Errorf("pg: get all counters failed error %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			pg.lg.ErrorCtx(ctx, "db: failed to close rows", zap.Error(closeErr))
		}
	}()

	for rows.Next() {
		c := models.Counter{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Value); err != nil {
			return nil, fmt.Errorf("pg: get all counters failed error %w", err)
		}
		counters = append(counters, c)
	}
	return counters, nil
}

func (pg *PG) Ping(ctx context.Context) error {
	return pg.db.PingContext(ctx)
}

func (pg *PG) Tx(ctx context.Context, fns ...func(ctx context.Context) error) error {
	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		if rberr := tx.Rollback(); rberr != nil {
			return errors.Join(fmt.Errorf("pg: rollback error %w", rberr), err)
		}

		return fmt.Errorf("pg: begin tx failed error %w", err)
	}

	for _, fn := range fns {
		if err := fn(ctx); err != nil {
			if rberr := tx.Rollback(); rberr != nil {
				return errors.Join(fmt.Errorf("pg: rollback error %w", rberr), err)
			}

			return fmt.Errorf("pg: tx failed error %w", err)
		}
	}

	return tx.Commit()
}
