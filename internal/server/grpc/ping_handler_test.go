package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mock "github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/storages"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNewPingHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	type args struct {
		strg *mock.MockStorage
		lg   *logging.ZapLogger
	}
	tests := []struct {
		name string
		args args
		want *PingHandler
	}{
		{
			name: "creates instance of PingHandler",
			args: args{
				strg: mockStorage,
				lg:   lg,
			},
			want: &PingHandler{
				strg: mockStorage,
				lg:   lg,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPingHandler(tt.args.strg, tt.args.lg)
			assert.Equal(t, tt.want.strg, got.strg)
			assert.Equal(t, tt.want.lg, got.lg)
		})
	}
}

func TestPingHandler_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	cfg := &config.Config{
		GRPCPort: "3200",
	}

	type fields struct {
		strg *mock.MockStorage
	}
	type args struct {
		params *emptypb.Empty
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(*fields)
		want    *emptypb.Empty
		wantErr bool
	}{
		{
			name:    "when ping failed",
			fields:  fields{},
			args:    args{params: &emptypb.Empty{}},
			wantErr: true,
			prepare: func(f *fields) {
				f.strg.EXPECT().Ping(gomock.Any()).Return(errors.New("ping failed"))
			},
		},
		{
			name:   "when ping ok",
			fields: fields{},
			args:   args{params: &emptypb.Empty{}},
			prepare: func(f *fields) {
				f.strg.EXPECT().Ping(gomock.Any()).Return(nil)
			},
			want: &emptypb.Empty{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.strg = mock.NewMockStorage(ctrl)
			tt.prepare(&tt.fields)

			handler := NewPingHandler(tt.fields.strg, lg)
			th := NewTestHandler(t, func(h *TestHandler) {
				h.PingHandler = *handler
			})

			RunTestServer(t, cfg, lg, th)

			client := NewTestClient(t, cfg)
			resp, err := client.Ping(context.Background(), &emptypb.Empty{})
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}
