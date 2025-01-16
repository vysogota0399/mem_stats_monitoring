package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

type tmock func(*gomock.Controller) storage.Storage

func okPingMock(ctrl *gomock.Controller) storage.Storage {
	m := mocks.NewMockDBAble(ctrl)
	m.EXPECT().Ping().Return(nil)
	return m
}

func errorPingMock(ctrl *gomock.Controller) storage.Storage {
	m := mocks.NewMockDBAble(ctrl)
	m.EXPECT().Ping().Return(errors.New(""))
	return m
}

func TestPingHandlerFunc(t *testing.T) {
	lg, err := logging.MustZapLogger(-1)
	assert.NoError(t, err)
	tests := []struct {
		name string
		strg tmock
		want int
	}{
		{
			name: "when DBAble storage and ping returns no errors, then status 200",
			strg: okPingMock,
			want: http.StatusOK,
		},
		{
			name: "when DBAble storage and ping retunes errors, then status 500",
			strg: errorPingMock,
			want: http.StatusInternalServerError,
		},
		{
			name: "when not DBAble storage, then status 406",
			strg: func(ctrl *gomock.Controller) storage.Storage { return nil },
			want: http.StatusNotAcceptable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			strg := tt.strg(ctrl)

			h := NewPingHandler(strg, lg)
			router := gin.Default()
			router.GET("/", h)
			r, err := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			if err != nil {
				assert.NoError(t, err)
			}

			router.ServeHTTP(w, r)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.want, resp.StatusCode)
		})
	}
}
