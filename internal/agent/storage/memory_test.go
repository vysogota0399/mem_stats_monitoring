package storage

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func NewMemoryStorageWithData(storage map[string]map[string]string, lg *logging.ZapLogger) *Memory {
	return &Memory{storage: storage, lg: lg}
}

func TestGet(t *testing.T) {
	tasks := []struct {
		name      string
		data      map[string]map[string]string
		mName     string
		mType     string
		want      string
		wantError error
	}{
		{
			name:      "when record found",
			data:      map[string]map[string]string{"counter": {"test": "1"}},
			mName:     "test",
			mType:     "counter",
			want:      "1",
			wantError: nil,
		},
		{
			name:      "when type not found",
			data:      map[string]map[string]string{"counter": {"test": "1"}},
			mName:     "test",
			mType:     "hist",
			want:      "",
			wantError: ErrNoRecords,
		},
		{
			name:      "when name not found",
			data:      map[string]map[string]string{"counter": {"test": "1"}},
			mName:     "supertest",
			mType:     "counter",
			want:      "",
			wantError: ErrNoRecords,
		},
		{
			name:      "when slice empty",
			data:      map[string]map[string]string{"counter": {"test": "1"}},
			mName:     "supertest",
			mType:     "counter",
			want:      "",
			wantError: ErrNoRecords,
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			lg, _ := logging.MustZapLogger(&config.Config{LogLevel: 1})
			storage := NewMemoryStorageWithData(tt.data, lg)
			m := &models.Metric{
				Type: tt.mType,
				Name: tt.mName,
			}
			err := storage.Get(m)
			assert.ErrorIs(t, err, tt.wantError)
			assert.Equal(t, tt.want, m.Value)
		})
	}
}

func TestSet(t *testing.T) {
	tasks := []struct {
		name string
		val  models.Metric
		data map[string]map[string]string
	}{
		{
			name: "push value to slice",
			val:  models.Metric{Name: "value", Value: "2", Type: "counter"},
			data: map[string]map[string]string{"counter": {"test": "2"}},
		},
		{
			name: "create new record",
			val:  models.Metric{Name: "test", Value: "1", Type: "counter"},
			data: make(map[string]map[string]string),
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			lg, _ := logging.MustZapLogger(&config.Config{LogLevel: 1})
			storage := NewMemoryStorageWithData(tt.data, lg)
			assert.NoError(t, storage.Set(context.Background(), &tt.val))

			actualValue := tt.data[tt.val.Type][tt.val.Name]
			assert.Equal(t, tt.val.Value, actualValue)
		})
	}
}

func TestNewMemoryStorage(t *testing.T) {
	type args struct {
		lg *logging.ZapLogger
	}
	tests := []struct {
		name string
		args args
		want *Memory
	}{
		{
			name: "successful memory storage creation",
			args: args{lg: nil},
			want: &Memory{storage: make(map[string]map[string]string), lg: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemoryStorage(tt.args.lg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemoryStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}
