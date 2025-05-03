package storages

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sync/atomic"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type MetricsRestorer struct {
	source io.ReadCloser
	lg     *logging.ZapLogger
	cfg    *config.Config
	strg   IStorageTarget
}

const rwfmode fs.FileMode = 0666

type IStorageTarget interface {
	CreateOrUpdate(ctx context.Context, mType, mName string, val any) error
	Tx(ctx context.Context, fns ...func(ctx context.Context) error) error
}

var _ Restorer = (*MetricsRestorer)(nil)
var _ IStorageTarget = (Storage)(nil)

type SourceBuilder interface {
	Source(cfg *config.Config) (io.ReadCloser, error)
}

func NewFileMetricsRestorer(
	cfg *config.Config,
	lg *logging.ZapLogger,
	strg IStorageTarget,
	srsb SourceBuilder,
) (*MetricsRestorer, error) {
	if cfg.FileStoragePath == "" {
		return nil, errors.New("restorer: file storage path is empty")
	}

	source, err := srsb.Source(cfg)
	if err != nil {
		return nil, err
	}

	return &MetricsRestorer{
		source: source,
		lg:     lg,
		cfg:    cfg,
		strg:   strg,
	}, nil
}

type persistentMetric struct {
	MType  string `json:"type"`
	MName  string `json:"name"`
	MValue any    `json:"value"`
}

func (r *MetricsRestorer) Call(ctx context.Context) error {
	r.lg.DebugCtx(ctx, "restore metrics from file")
	scanner := bufio.NewScanner(r.source)
	var sucCntr int64

	defer func(i *int64) {
		r.lg.DebugCtx(ctx, "storage restored", zap.Int64("succeeded", *i))
		if err := r.source.Close(); err != nil {
			r.lg.ErrorCtx(ctx, "failed to close file", zap.Error(err))
		}
	}(&sucCntr)

	operations := make([]func(ctx context.Context) error, 0)

	i := 0
	for scanner.Scan() {
		i++
		var record persistentMetric
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return fmt.Errorf("restore: unmarshal record failed %w, token: %s, line: %d", err, scanner.Text(), i)
		}

		operations = append(operations, func(ctx context.Context) error {
			atomic.AddInt64(&sucCntr, 1)
			r.lg.DebugCtx(ctx, "save new record to storage", zap.Any("record", record))

			switch record.MType {
			case models.CounterType:
				return r.strg.CreateOrUpdate(ctx, record.MType, record.MName, int64(record.MValue.(float64)))
			case models.GaugeType:
				return r.strg.CreateOrUpdate(ctx, record.MType, record.MName, record.MValue.(float64))
			default:
				return fmt.Errorf("restore: unknown metric type %s", record.MType)
			}
		})
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("restorer: failed to scan file %w", err)
	}

	if len(operations) == 0 {
		r.lg.DebugCtx(ctx, "no records to restore")
		return nil
	}

	if err := r.strg.Tx(ctx, operations...); err != nil {
		return fmt.Errorf("restore: failed to save records %w", err)
	}

	return nil
}
