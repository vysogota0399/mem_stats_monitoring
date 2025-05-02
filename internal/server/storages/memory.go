package storages

import (
	"context"
	"fmt"
	"sync"

	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

type Memory struct {
	storage map[string]map[string]any
	mutex   sync.RWMutex
	lg      *logging.ZapLogger
}

func NewMemory(lg *logging.ZapLogger) *Memory {
	return &Memory{
		storage: make(map[string]map[string]any),
		mutex:   sync.RWMutex{},
		lg:      lg,
	}
}

// CreateOrUpdate implements Storage. CreateOrUpdate is used to create or update a metric.
// It should not lock the mutex, because it will be called from Tx.
func (m *Memory) CreateOrUpdate(ctx context.Context, mType, mName string, val any) error {
	if ctx.Value(txInProgressKey) == nil {
		m.mutex.Lock()
		defer m.mutex.Unlock()
	}

	m.lg.DebugCtx(ctx, "create or update metric", zap.String("mType", mType), zap.String("mName", mName))
	mTypeStorage, ok := m.storage[mType]
	if !ok {
		mTypeStorage = make(map[string]any)
		m.storage[mType] = mTypeStorage
	}

	mTypeStorage[mName] = val
	return nil
}

func (m *Memory) GetGauge(ctx context.Context, record *models.Gauge) error {
	m.lg.DebugCtx(ctx, "get gauge", zap.Any("record", record))
	defer m.lg.DebugCtx(ctx, "get gauge done")

	if record.Name == "" {
		return fmt.Errorf("memory: gauge name is empty")
	}

	if ctx.Value(txInProgressKey) == nil {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
	}

	m.lg.DebugCtx(ctx, "get metric", zap.String("mType", models.GaugeType), zap.String("mName", record.Name))
	types, ok := m.storage[models.GaugeType]
	if !ok {
		return fmt.Errorf("memory: no records for gauge %w", ErrNoRecords)
	}

	val, ok := types[record.Name]
	if !ok {
		return fmt.Errorf("memory: no records for gauge with name %s %w", record.Name, ErrNoRecords)
	}

	v, ok := val.(float64)
	if !ok {
		return fmt.Errorf("memory: invalid gauge value type %T, record: %+v, value: %+v", v, record, val)
	}

	record.Value = v

	return nil
}

func (m *Memory) GetCounter(ctx context.Context, record *models.Counter) error {
	m.lg.DebugCtx(ctx, "get counter", zap.Any("record", record))
	defer m.lg.DebugCtx(ctx, "get counter done")

	if record.Name == "" {
		return fmt.Errorf("memory: counter name is empty")
	}

	if ctx.Value(txInProgressKey) == nil {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
	}

	m.lg.DebugCtx(ctx, "get metric", zap.String("mType", models.CounterType), zap.String("mName", record.Name))
	types, ok := m.storage[models.CounterType]
	if !ok {
		return fmt.Errorf("memory: no records for counter %w", ErrNoRecords)
	}

	val, ok := types[record.Name]
	if !ok {
		return fmt.Errorf("memory: no records for counter with name %s %w", record.Name, ErrNoRecords)
	}

	switch v := val.(type) {
	case int64:
		record.Value = v
	case int:
		record.Value = int64(v)
	default:
		return fmt.Errorf("memory: invalid counter value type %T, record: %+v, value: %+v", v, record, val)
	}

	return nil
}

func (m *Memory) GetGauges(ctx context.Context) ([]models.Gauge, error) {
	if ctx.Value(txInProgressKey) == nil {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
	}

	gauges := make([]models.Gauge, 0)
	for mName, val := range m.storage[models.GaugeType] {
		gauges = append(gauges, models.Gauge{
			Name:  mName,
			Value: val.(float64),
		})
	}

	return gauges, nil
}

func (m *Memory) GetCounters(ctx context.Context) ([]models.Counter, error) {
	if ctx.Value(txInProgressKey) == nil {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
	}

	counters := make([]models.Counter, 0)
	for mName, val := range m.storage[models.CounterType] {
		counters = append(counters, models.Counter{
			Name:  mName,
			Value: val.(int64),
		})
	}

	return counters, nil
}

func (m *Memory) Ping(ctx context.Context) error {
	return nil
}

type txInProgress string

const (
	txInProgressKey txInProgress = "txInProgress"
)

func (m *Memory) Tx(ctx context.Context, fns ...func(ctx context.Context) error) error {
	if ctx.Value(txInProgressKey) != nil {
		return fmt.Errorf("memory: tx already in progress")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	txCtx := context.WithValue(ctx, txInProgressKey, struct{}{})

	m.lg.DebugCtx(txCtx, "[LOCK] tx started", zap.Int("operations count", len(fns)))

	for _, fn := range fns {
		m.lg.DebugCtx(txCtx, "tx operation started")
		if err := fn(txCtx); err != nil {
			return fmt.Errorf("memory: tx failed error %w", err)
		}
	}

	m.lg.DebugCtx(txCtx, "[UNLOCK] tx succeeded")
	return nil
}
