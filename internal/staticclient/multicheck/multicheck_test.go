package multicheck

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"go.uber.org/zap"
)

// ReadWriteCloserBuffer implements io.ReadCloser interface
type ReadWriteCloserBuffer struct {
	*bytes.Buffer
}

func (b ReadWriteCloserBuffer) Close() error {
	return nil
}

func TestCall(t *testing.T) {
	var lg *logging.ZapLogger
	{
		var err error
		lg, err = logging.MustZapLogger(zap.DebugLevel)
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		source    io.ReadCloser
		wantErr   bool
		wantCheck *MultiCheck
	}{
		{
			name: "successful creation with valid yaml",
			source: ReadWriteCloserBuffer{
				Buffer: bytes.NewBufferString(`
analyzers:
  - name: staticcheck
    checks:
      - QF1010
      - QF1003
`),
			},
			wantErr: false,
			wantCheck: &MultiCheck{
				ExtraAnalyzers: []Analyzer{
					{
						Name:   "staticcheck",
						Checks: []string{"QF1010", "QF1003"},
					},
				},
			},
		},
		{
			name: "invalid yaml",
			source: ReadWriteCloserBuffer{
				Buffer: bytes.NewBufferString("invalid: yaml: content: {"),
			},
			wantErr: true,
			wantCheck: &MultiCheck{
				ExtraAnalyzers: []Analyzer{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new MultiCheck with the test source
			mcheck := &MultiCheck{
				lg:     lg,
				source: tt.source,
				skip:   true,
			}

			// Test
			err := mcheck.Call()

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantCheck != nil {
					assert.Equal(t, tt.wantCheck.ExtraAnalyzers, mcheck.ExtraAnalyzers)
				}
			}

			// Cleanup
			err = tt.source.Close()
			assert.NoError(t, err)
		})
	}
}
