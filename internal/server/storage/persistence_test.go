package storage

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
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
