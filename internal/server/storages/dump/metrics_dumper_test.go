package dump

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

func TestNewMetricsDumper(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	dumper := NewMetricsDumper(lg, &config.Config{})

	// Verify that the dumper is initialized with the correct logger
	assert.Equal(t, lg, dumper.lg)

	// Verify that the publisher and subscriber are initialized
	assert.NotNil(t, dumper.pb)
	assert.NotNil(t, dumper.sb)

	// Verify that the publisher has the correct logger
	assert.Equal(t, lg, dumper.pb.lg)

	// Verify that the subscriber has the correct logger
	assert.Equal(t, lg, dumper.sb.lg)
}

type WriteCloserTarget struct {
	*bytes.Buffer
}

func (w *WriteCloserTarget) Close() error {
	return nil
}

func TestMetricsDumper_StartSync(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	target := &WriteCloserTarget{Buffer: &bytes.Buffer{}}

	dumper := NewMetricsDumper(lg, &config.Config{})
	dumper.start(context.Background(), target)

	expected := make([]DumpMessage, 0, 10)
	for i := 0; i < cap(expected); i++ {
		expected = append(expected, DumpMessage{MName: fmt.Sprintf("name_%d", i)})
	}

	ctx := context.Background()

	for _, m := range expected {
		dumper.Dump(ctx, m)
	}
	dumper.Stop(ctx)

	scanner := bufio.NewScanner(target.Buffer)
	actual := make([]DumpMessage, 0, len(expected))
	for scanner.Scan() {
		var record DumpMessage
		err := json.Unmarshal(scanner.Bytes(), &record)
		assert.NoError(t, err)
		actual = append(actual, record)
	}

	assert.Equal(t, expected, actual)
}

func TestMetricsDumper_StartAsync(t *testing.T) {
	lg, err := logging.MustZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	target := &WriteCloserTarget{Buffer: &bytes.Buffer{}}

	var waitSeconds int64 = 1
	dumper := NewMetricsDumper(lg, &config.Config{StoreInterval: waitSeconds * 2})

	dumper.start(context.Background(), target)

	expected := make([]DumpMessage, 0, 10)
	for i := 0; i < cap(expected); i++ {
		expected = append(expected, DumpMessage{})
	}

	ctx := context.Background()

	for _, m := range expected {
		dumper.Dump(ctx, m)
	}

	time.Sleep(time.Duration(waitSeconds) * time.Second)

	dumper.Stop(ctx)

	scanner := bufio.NewScanner(target.Buffer)
	actual := make([]DumpMessage, 0, len(expected))
	for scanner.Scan() {
		var record DumpMessage
		err := json.Unmarshal(scanner.Bytes(), &record)
		assert.NoError(t, err)
		actual = append(actual, record)
	}

	assert.Equal(t, expected, actual)
}
