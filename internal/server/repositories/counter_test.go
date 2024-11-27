package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
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
				utils.InitLogger("[test]"),
			),
		},
		{
			record:      models.Counter{Value: 2, Name: "test"},
			wantsRecord: models.Counter{Value: 3, Name: "test"},
			name:        "when storage has record then create add one",
			storage: storage.NewMemStorageWithData(
				map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
				utils.InitLogger("[test]"),
			),
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			subject := NewCounter(tt.storage)
			record, err := subject.Craete(tt.record)
			assert.NoError(t, err)
			assert.Equal(t, record, tt.wantsRecord)
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
				utils.InitLogger("[test]"),
			),
		},
		{
			searchName:  "test2",
			wantsRecord: nil,
			wantsError:  storage.ErrNoRecords,
			storage: storage.NewMemStorageWithData(
				map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
				utils.InitLogger("[test]"),
			),
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			subject := NewCounter(tt.storage)
			record, err := subject.Last(tt.searchName)
			assert.Equal(t, record, tt.wantsRecord)
			assert.Equal(t, err, tt.wantsError)
		})
	}
}
