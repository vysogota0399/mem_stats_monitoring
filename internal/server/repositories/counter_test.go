package repositories

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestCreate(t *testing.T) {
	tasks := []struct {
		record      models.Counter
		wantsRecord models.Counter
		name        string
		storage     storage.Storage
	}{
		{
			record:      models.Counter{Value: 1, Name: "test"},
			wantsRecord: models.Counter{Value: 1, Name: "test"},
			name:        "when storage empty create then record to state",
			storage: storage.NewMemStorageWithData(
				map[string]map[string][]string{},
			),
		},
		{
			record:      models.Counter{Value: 2, Name: "test"},
			wantsRecord: models.Counter{Value: 3, Name: "test"},
			name:        "when storage has record then create add one",
			storage: storage.NewMemStorageWithData(
				map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
			),
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			subject := NewCounter(tt.storage, lg)
			record, err := subject.Create(context.Background(), &tt.record)
			assert.NoError(t, err)
			assert.NotNil(t, record)
			assert.Equal(t, tt.wantsRecord, *record)
		})
	}
}

func TestLast(t *testing.T) {
	tasks := []struct {
		wantsRecord *models.Counter
		wantsError  error
		name        string
		storage     storage.Storage
		searchName  string
	}{
		{
			searchName:  "test",
			wantsRecord: &models.Counter{Value: 1, Name: "test"},
			wantsError:  nil,
			name:        "when storage has record then returns error",
			storage: storage.NewMemStorageWithData(
				map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
			),
		},
		{
			searchName:  "test2",
			wantsRecord: nil,
			wantsError:  storage.ErrNoRecords,
			storage: storage.NewMemStorageWithData(
				map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
			),
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			subject := NewCounter(tt.storage, lg)
			record, err := subject.Last(context.Background(), tt.searchName)
			assert.Equal(t, record, tt.wantsRecord)
			assert.Equal(t, err, tt.wantsError)
		})
	}
}

func TestCounter_All(t *testing.T) {
	type fields struct {
		storage storage.Storage
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string][]models.Counter
	}{
		{
			name: "when storage has values then returns collection",
			fields: fields{
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{
						"counter": {
							"fiz": []string{`{"value": 0, "name": "fiz"}`},
							"baz": []string{`{"value": 0, "name": "baz"}`},
						},
					},
				),
			},
			want: map[string][]models.Counter{
				"fiz": {models.Counter{Name: "fiz", Value: 0}},
				"baz": {models.Counter{Name: "baz", Value: 0}},
			},
		},
		{
			name: "when storage has no values then returns empty collection",
			fields: fields{
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{},
				),
			},
			want: map[string][]models.Counter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			c := NewCounter(tt.fields.storage, lg)
			actual := c.All()
			assert.Equal(t, tt.want, actual)
		})
	}
}

func TestCounter_SaveCollection(t *testing.T) {
	type fields struct {
		storage    storage.Storage
		collection []models.Counter
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
				collection: []models.Counter{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			g := &Counter{
				storage: tt.fields.storage,
				lg:      lg,
			}
			_, err = g.SaveCollection(context.Background(), tt.fields.collection)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestCounter_Create(t *testing.T) {
	type fields struct {
		storage storage.Storage
		record  models.Counter
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
				record:  models.Counter{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			g := &Counter{
				storage: tt.fields.storage,
				lg:      lg,
			}
			_, err = g.Create(context.Background(), &tt.fields.record)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestCounter_SearchByName(t *testing.T) {
	type fields struct {
		storage storage.Storage
	}
	type args struct {
		names []string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]models.Counter
		args   args
	}{
		{
			name: "when storage has values then returns collection",
			args: args{names: []string{"fiz"}},
			fields: fields{
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{
						"counter": {
							"fiz": []string{`{"value": 0, "name": "fiz"}`},
							"baz": []string{`{"value": 0, "name": "baz"}`},
						},
					},
				),
			},
			want: map[string]models.Counter{
				"fiz": {Name: "fiz", Value: 0},
			},
		},
		{
			name: "when storage has no values then returns empty collection",
			fields: fields{
				storage: storage.NewMemStorageWithData(
					map[string]map[string][]string{},
				),
			},
			want: map[string]models.Counter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logging.MustZapLogger(-1)
			assert.NoError(t, err)
			c := &Counter{
				storage: tt.fields.storage,
				lg:      lg,
			}
			got, err := c.SearchByName(context.Background(), tt.args.names)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
