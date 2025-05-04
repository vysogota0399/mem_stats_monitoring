package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

type PGConnectionOpener struct {
	db  *sql.DB
	lg  *logging.ZapLogger
	dsn string
}

func NewPGConnectionOpener(dsn string, lg *logging.ZapLogger) (*PGConnectionOpener, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PGConnectionOpener{
		db:  db,
		lg:  lg,
		dsn: dsn,
	}, nil
}

func (p *PGConnectionOpener) Close() error {
	return p.db.Close()
}

func (p *PGConnectionOpener) Get(to *models.Metric) error {
	query := `SELECT value FROM metrics WHERE type = $1 AND name = $2`
	row := p.db.QueryRow(query, to.Type, to.Name)

	var value string
	if err := row.Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("metric not found: type=%s, name=%s", to.Type, to.Name)
		}
		return fmt.Errorf("failed to get metric: %w", err)
	}

	to.Value = value
	return nil
}

func (p *PGConnectionOpener) Set(ctx context.Context, m *models.Metric) error {
	query := `
		INSERT INTO metrics (type, name, value)
		VALUES ($1, $2, $3)
		ON CONFLICT (type, name) DO UPDATE
		SET value = EXCLUDED.value
	`

	_, err := p.db.ExecContext(ctx, query, m.Type, m.Name, m.Value)
	if err != nil {
		return fmt.Errorf("failed to set metric: %w", err)
	}

	return nil
}
