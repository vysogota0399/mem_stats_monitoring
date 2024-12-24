package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshaler(t *testing.T) {
	type fields struct {
		Value float64
		Name  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "float 737731.0",
			fields: fields{Value: 737731., Name: "test"},
			want:   `{"value": 737731.0, "name": "test"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Gauge{
				Value: tt.fields.Value,
				Name:  tt.fields.Name,
			}

			marshaled, err := json.Marshal(g)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.want, string(marshaled))
		})
	}
}
