package repositories

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
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
			got, err := g.Last(context.Background(), tt.args.mName)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestGauge_All(t *testing.T) {
	type fields struct {
		storage storage.Storage
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string][]models.Gauge
	}{
		{
			name: "return all records",
			fields: fields{
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{
						"gauge": {
							"testSetGet241": []string{`{"value": 120400.951, "name": "testSetGet241"}`},
						},
					},
				),
			},
			want: map[string][]models.Gauge{
				"testSetGet241": []models.Gauge{{Value: 120400.951, Name: "testSetGet241"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Gauge{
				storage: tt.fields.storage,
			}
			got := g.All()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGauge_SaveCollection(t *testing.T) {
	type fields struct {
		storage    storage.Storage
		collection []models.Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "when inmemory storage",
			fields: fields{
				storage:    storage.NewMemory(),
				collection: []models.Gauge{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				storage: tt.fields.storage,
			}
			_, err := g.SaveCollection(context.Background(), tt.fields.collection)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestGauge_Create(t *testing.T) {
	type fields struct {
		storage storage.Storage
		record  models.Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "when inmemory storage",
			fields: fields{
				storage: storage.NewMemory(),
				record:  models.Gauge{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Gauge{
				storage: tt.fields.storage,
			}
			_, err := g.Create(context.Background(), &tt.fields.record)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
