package storages

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNewMemory(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	memory := NewMemory(lg)

	// Check that storage is initialized as an empty map
	assert.NotNil(t, memory.storage)
	assert.Empty(t, memory.storage)

	// Check that logger is set correctly
	assert.Equal(t, lg, memory.lg)
}

func TestMemory_CreateOrUpdate(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		mType   string
		mName   string
		val     any
		wantErr bool
	}{
		{
			name:    "create new gauge",
			mType:   models.GaugeType,
			mName:   "test_gauge",
			val:     42.0,
			wantErr: false,
		},
		{
			name:    "update existing gauge",
			mType:   models.GaugeType,
			mName:   "test_gauge",
			val:     84.0,
			wantErr: false,
		},
		{
			name:    "create new counter",
			mType:   models.CounterType,
			mName:   "test_counter",
			val:     int64(42),
			wantErr: false,
		},
		{
			name:    "update existing counter",
			mType:   models.CounterType,
			mName:   "test_counter",
			val:     int64(84),
			wantErr: false,
		},
		{
			name:    "invalid metric type",
			mType:   "invalid_type",
			mName:   "test",
			val:     42.0,
			wantErr: false, // The function doesn't validate metric types
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMemory(lg)
			err := m.CreateOrUpdate(ctx, tt.mType, tt.mName, tt.val)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				mTypeStorage, exists := m.storage[tt.mType]
				assert.True(t, exists)

				val, exists := mTypeStorage[tt.mName]
				assert.True(t, exists)
				assert.Equal(t, tt.val, val)
			}
		})
	}
}

func TestMemory_GetGauge(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name      string
		preload   map[string]map[string]any
		record    *models.Gauge
		wantValue float64
		wantErr   bool
	}{
		{
			name: "successful get existing gauge",
			preload: map[string]map[string]any{
				models.GaugeType: {
					"test_gauge": 42.0,
				},
			},
			record: &models.Gauge{
				Name: "test_gauge",
			},
			wantValue: 42.0,
			wantErr:   false,
		},
		{
			name: "empty gauge name",
			preload: map[string]map[string]any{
				models.GaugeType: {
					"test_gauge": 42.0,
				},
			},
			record: &models.Gauge{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "gauge type not found",
			preload: map[string]map[string]any{
				models.CounterType: {
					"test_counter": int64(42),
				},
			},
			record: &models.Gauge{
				Name: "test_gauge",
			},
			wantErr: true,
		},
		{
			name: "gauge name not found",
			preload: map[string]map[string]any{
				models.GaugeType: {
					"other_gauge": 42.0,
				},
			},
			record: &models.Gauge{
				Name: "test_gauge",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				storage: tt.preload,
				mutex:   sync.RWMutex{},
				lg:      lg,
			}

			err := m.GetGauge(ctx, tt.record)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, tt.record.Value)
			}
		})
	}
}

func TestMemory_GetCounter(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name      string
		preload   map[string]map[string]any
		record    *models.Counter
		wantValue int64
		wantErr   bool
	}{
		{
			name: "successful get existing counter",
			preload: map[string]map[string]any{
				models.CounterType: {
					"test_counter": int64(42),
				},
			},
			record: &models.Counter{
				Name: "test_counter",
			},
			wantValue: 42,
			wantErr:   false,
		},
		{
			name: "empty counter name",
			preload: map[string]map[string]any{
				models.CounterType: {
					"test_counter": int64(42),
				},
			},
			record: &models.Counter{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "counter type not found",
			preload: map[string]map[string]any{
				models.GaugeType: {
					"test_gauge": 42.0,
				},
			},
			record: &models.Counter{
				Name: "test_counter",
			},
			wantErr: true,
		},
		{
			name: "counter name not found",
			preload: map[string]map[string]any{
				models.CounterType: {
					"other_counter": int64(42),
				},
			},
			record: &models.Counter{
				Name: "test_counter",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				storage: tt.preload,
				mutex:   sync.RWMutex{},
				lg:      lg,
			}

			err := m.GetCounter(ctx, tt.record)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, tt.record.Value)
			}
		})
	}
}

func TestMemory_GetGauges(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		preload map[string]map[string]any
		want    []models.Gauge
		wantErr bool
	}{
		{
			name: "get all gauges",
			preload: map[string]map[string]any{
				models.GaugeType: {
					"gauge1": 42.0,
					"gauge2": 84.0,
				},
				models.CounterType: {
					"counter1": int64(42),
				},
			},
			want: []models.Gauge{
				{Name: "gauge1", Value: 42.0},
				{Name: "gauge2", Value: 84.0},
			},
			wantErr: false,
		},
		{
			name: "no gauges",
			preload: map[string]map[string]any{
				models.CounterType: {
					"counter1": int64(42),
				},
			},
			want:    []models.Gauge{},
			wantErr: false,
		},
		{
			name:    "empty storage",
			preload: map[string]map[string]any{},
			want:    []models.Gauge{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				storage: tt.preload,
				mutex:   sync.RWMutex{},
				lg:      lg,
			}

			got, err := m.GetGauges(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.want, got)
			}
		})
	}
}

func TestMemory_GetCounters(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		preload map[string]map[string]any
		want    []models.Counter
		wantErr bool
	}{
		{
			name: "get all counters",
			preload: map[string]map[string]any{
				models.CounterType: {
					"counter1": int64(42),
					"counter2": int64(84),
				},
				models.GaugeType: {
					"gauge1": 42.0,
				},
			},
			want: []models.Counter{
				{Name: "counter1", Value: 42},
				{Name: "counter2", Value: 84},
			},
			wantErr: false,
		},
		{
			name: "no counters",
			preload: map[string]map[string]any{
				models.GaugeType: {
					"gauge1": 42.0,
				},
			},
			want:    []models.Counter{},
			wantErr: false,
		},
		{
			name:    "empty storage",
			preload: map[string]map[string]any{},
			want:    []models.Counter{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				storage: tt.preload,
				mutex:   sync.RWMutex{},
				lg:      lg,
			}

			got, err := m.GetCounters(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.want, got)
			}
		})
	}
}

func TestMemory_Ping(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		preload map[string]map[string]any
		wantErr bool
	}{
		{
			name:    "ping with empty storage",
			preload: map[string]map[string]any{},
			wantErr: false,
		},
		{
			name: "ping with populated storage",
			preload: map[string]map[string]any{
				models.GaugeType: {
					"gauge1": 42.0,
				},
				models.CounterType: {
					"counter1": int64(42),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				storage: tt.preload,
				mutex:   sync.RWMutex{},
				lg:      lg,
			}

			err := m.Ping(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMemory_Tx(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		preload map[string]map[string]any
		fns     []func(ctx context.Context) error
		wantErr bool
	}{
		{
			name: "successful transaction with multiple operations",
			preload: map[string]map[string]any{
				models.GaugeType: {
					"gauge1": 42.0,
				},
			},
			fns: []func(ctx context.Context) error{
				func(ctx context.Context) error {
					return nil
				},
				func(ctx context.Context) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name:    "transaction with error in first operation",
			preload: map[string]map[string]any{},
			fns: []func(ctx context.Context) error{
				func(ctx context.Context) error {
					return errors.New("first operation failed")
				},
				func(ctx context.Context) error {
					return nil
				},
			},
			wantErr: true,
		},
		{
			name:    "transaction with error in second operation",
			preload: map[string]map[string]any{},
			fns: []func(ctx context.Context) error{
				func(ctx context.Context) error {
					return nil
				},
				func(ctx context.Context) error {
					return errors.New("second operation failed")
				},
			},
			wantErr: true,
		},
		{
			name:    "empty transaction",
			preload: map[string]map[string]any{},
			fns:     []func(ctx context.Context) error{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Memory{
				storage: tt.preload,
				mutex:   sync.RWMutex{},
				lg:      lg,
			}

			err := m.Tx(ctx, tt.fns...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
