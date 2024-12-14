package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

func TestGauge_Last(t *testing.T) {
	type args struct {
		storage storage.Storage
		mName   string
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Gauge
		wantErr error
	}{
		{
			name: "when storage has record then returns record",
			args: args{
				mName: "testSetGet241",
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{
						"gauge": {
							"testSetGet241": []string{`{"value": 120400.951, "name": "testSetGet241"}`},
						},
					},
				),
			},
			want:    &models.Gauge{Value: 120400.951, Name: "testSetGet241"},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGauge(tt.args.storage)
			got, err := g.Last(tt.args.mName)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
