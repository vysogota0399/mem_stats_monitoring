package server

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type MockHandler struct {
}

func (h *MockHandler) Registrate() (Route, error) {
	return Route{
		Path:    "/",
		Method:  "GET",
		Handler: func(c *gin.Context) {},
	}, nil
}

func TestNewRouter(t *testing.T) {
	type args struct {
		handlers []Handler
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "when create router",
			args: args{
				handlers: []Handler{
					&MockHandler{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				l fx.Lifecycle
				s fx.Shutdowner
			)
			app := fxtest.New(
				t,
				fx.Populate(&l, &s),
			)
			got := NewRouter(tt.args.handlers, l, nil, &config.Config{}, nil)
			assert.NotNil(t, got)
			err := app.Start(context.Background())
			assert.NoError(t, err)
		})
	}
}
