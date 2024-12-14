package storage

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

func TestLast(t *testing.T) {
	tasks := []struct {
		name       string
		data       map[string]map[string][]string
		mName      string
		mType      string
		wantRecord string
		wantError  error
	}{
		{
			name:       "when record found",
			data:       map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
			mName:      "test",
			mType:      "counter",
			wantRecord: `{"value": 1, "name": "test"}`,
			wantError:  nil,
		},
		{
			name:       "when type not found",
			data:       map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
			mName:      "test",
			mType:      "hist",
			wantRecord: ``,
			wantError:  ErrNoRecords,
		},
		{
			name:       "when name not found",
			data:       map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
			mName:      "supertest",
			mType:      "counter",
			wantRecord: ``,
			wantError:  ErrNoRecords,
		},
		{
			name:       "when slice empty",
			data:       map[string]map[string][]string{"counter": {"test": []string{}}},
			mName:      "supertest",
			mType:      "counter",
			wantRecord: ``,
			wantError:  ErrNoRecords,
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorageWithData(tt.data, utils.InitLogger("[test]"))
			val, err := storage.Last(tt.mType, tt.mName)
			assert.ErrorIs(t, tt.wantError, err)
			assert.Equal(t, tt.wantRecord, val)
		})
	}
}

func TestPush(t *testing.T) {
	type record struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tasks := []struct {
		name string
		val  record
		data map[string]map[string][]string
	}{
		{
			name: "push value to slice",
			val:  record{Name: "value", Value: 9999},
			data: map[string]map[string][]string{"counter": {"test": []string{`{"value": 1, "name": "test"}`}}},
		},
		{
			name: "create new record",
			val:  record{Name: "test", Value: 10},
			data: map[string]map[string][]string{},
		},
	}

	for _, tt := range tasks {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorageWithData(tt.data, utils.InitLogger("[test]"))
			err := storage.Push("counter", tt.val.Name, tt.val)
			assert.NoError(t, err)

			expectedVal, err := json.Marshal(tt.val)
			assert.NoError(t, err)
			records := tt.data["counter"][tt.val.Name]
			expected := string(expectedVal)
			actual := records[len(records)-1]
			assert.JSONEq(t, expected, actual)
		})
	}
}
