package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils"
)

func TestNewRootHandler(t *testing.T) {
	type args struct {
		storage storage.Storage
	}
	tests := []struct {
		name string
		args args
		want gin.HandlerFunc
	}{
		{
			name: "returns status 200",
			args: args{
				storage: storage.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			router.LoadHTMLGlob("../templates/*.tmpl")
			handler := NewRootHandler(tt.args.storage, utils.InitLogger("[test]"))
			router.GET("/", handler)

			r, err := http.NewRequestWithContext(context.TODO(), "GET", "/", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			if err != nil {
				assert.NoError(t, err)
			}

			router.ServeHTTP(w, r)
			response := w.Result()
			defer response.Body.Close()

			assert.Equal(t, http.StatusOK, response.StatusCode)
		})
	}
}
