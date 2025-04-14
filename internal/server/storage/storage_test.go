package storage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want Storage
	}{
		{
			name: "when called returns new memory instance",
			want: &Memory{
				storage: make(map[string]map[string][]string),
				mutex:   sync.Mutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New()
			assert.NotNil(t, got)
			assert.IsType(t, &Memory{}, got)
		})
	}
}
