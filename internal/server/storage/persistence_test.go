package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"golang.org/x/sync/errgroup"
)

func TestPersistentMemory_restore(t *testing.T) {
	type fields struct {
		dSource dataSource
	}
	tests := []struct {
		name    string
		fields  fields
		want    []persistentMetric
		wantErr bool
	}{
		{
			name: "when restore any metrics",
			fields: fields{
				dSource: func(m *PersistentMemory) (io.ReadCloser, error) {
					return io.NopCloser(
						strings.NewReader(`
{"name":"GetSet1","type":"counter","value":1}
{"name":"GetSet2","type":"counter","value":2}
					`)), nil
				},
			},
			want: []persistentMetric{
				{
					MName:  "GetSet1",
					MType:  "counter",
					MValue: 1.,
				},
				{
					MName:  "GetSet2",
					MType:  "counter",
					MValue: 2.,
				},
			},
			wantErr: false,
		},
		{
			name: "when restore none metrics",
			want: make([]persistentMetric, 0),
			fields: fields{
				dSource: func(m *PersistentMemory) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			m := &PersistentMemory{
				Memory:  *NewMemory(),
				ctx:     context.Background(),
				lg:      lg,
				dSource: tt.fields.dSource,
			}
			got, err := m.restore()
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestPersistentMemory_Push(t *testing.T) {
	type fields struct {
		tmpfile *os.File
	}
	type args struct {
		c    func(f *fields) (config.Config, error)
		errg *errgroup.Group
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    *PersistentMemory
		wantErr bool
	}{
		{
			name: "with restore",
			args: args{
				errg: &errgroup.Group{},
				c: func(f *fields) (config.Config, error) {
					file, err := os.CreateTemp("", "*")
					if err != nil {
						return config.Config{}, fmt.Errorf("persistence_test: create temp file error %w", err)
					}

					f.tmpfile = file
					return config.Config{Restore: true, FileStoragePath: f.tmpfile.Name()}, nil
				},
			},
		},
		{
			name: "without restore",
			args: args{
				errg: &errgroup.Group{},
				c: func(f *fields) (config.Config, error) {
					file, err := os.CreateTemp("", "*")
					if err != nil {
						return config.Config{}, fmt.Errorf("persistence_test: create temp file error %w", err)
					}

					f.tmpfile = file
					return config.Config{Restore: true, FileStoragePath: f.tmpfile.Name()}, nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)

			cfg, err := tt.args.c(&tt.fields)
			assert.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			s, err := NewFilePersistentMemory(ctx, cfg, tt.args.errg, lg)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.NotNil(t, s)
			assert.NoError(t, s.Push(models.CounterType, "", 1))

			cancel()
			assert.NoError(t, tt.args.errg.Wait())
		})
	}
}
