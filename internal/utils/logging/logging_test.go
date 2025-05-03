package logging

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"go.uber.org/zap"
)

func TestZapLogger(t *testing.T) {
	tests := []struct {
		name string
		do   func(ctx context.Context, l *ZapLogger, val int64, wg *sync.WaitGroup)
	}{
		{
			name: "when info",
			do: func(ctx context.Context, l *ZapLogger, val int64, wg *sync.WaitGroup) {
				defer wg.Done()

				ctx = l.WithContextFields(ctx, zap.Int64(fmt.Sprintf("_rand_%d", val), val))
				l.InfoCtx(ctx, "", zap.Int64("val", val))
			},
		},
		{
			name: "when debug",
			do: func(ctx context.Context, l *ZapLogger, val int64, wg *sync.WaitGroup) {
				defer wg.Done()

				l.InfoCtx(ctx, "", zap.Int64("val", val))
			},
		},
		{
			name: "when warn",
			do: func(ctx context.Context, l *ZapLogger, val int64, wg *sync.WaitGroup) {
				defer wg.Done()

				l.InfoCtx(ctx, "", zap.Int64("val", val))
			},
		},
		{
			name: "when error",
			do: func(ctx context.Context, l *ZapLogger, val int64, wg *sync.WaitGroup) {
				defer wg.Done()

				l.InfoCtx(ctx, "", zap.Int64("val", val))
			},
		},
		{
			name: "when fatal",
			do: func(ctx context.Context, l *ZapLogger, val int64, wg *sync.WaitGroup) {
				defer wg.Done()

				l.InfoCtx(ctx, "", zap.Int64("val", val))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := NewZapLogger(&config.Config{LogLevel: -1})
			assert.NoError(t, err)

			wg := &sync.WaitGroup{}
			ctx := context.Background()
			n := 100
			m := 4
			wg.Add(m * n)
			for i := range n {
				ctx = lg.WithContextFields(ctx, zap.Int(fmt.Sprintf("rand_%d", i), rand.Int()))
				inti := int64(i)
				for range m {
					go tt.do(ctx, lg, int64(inti), wg)
				}
			}

			wg.Wait()
		})
	}
}
