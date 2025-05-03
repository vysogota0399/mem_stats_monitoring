package grpc

import (
	"log"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vysogota0399/mem_stats_monitoring/internal/mocks/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/config"

	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNewIndexHandler(t *testing.T) {
	type args struct {
		gaugeRepository   IMetricsGaugeRepository
		counterRepository IMetricsCounterRepository
		lg                *logging.ZapLogger
	}
	tests := []struct {
		name string
		args args
		want *IndexHandler
	}{
		{
			name: "creates instance of IndexHandler",
			args: args{
				gaugeRepository:   nil,
				counterRepository: nil,
				lg:                nil,
			},
			want: &IndexHandler{
				gaugeRepository:   nil,
				counterRepository: nil,
				lg:                nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewIndexHandler(tt.args.gaugeRepository, tt.args.counterRepository, tt.args.lg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIndexHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIndexHandler_Index(t *testing.T) {
	type fields struct {
		gaugeRepository   *repositories.MockIGaugeRepository
		counterRepository *repositories.MockICounterRepository
	}
	type args struct {
		params *emptypb.Empty
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		prepare func(*fields)
		want    *metrics.IndexResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}

	cfg := &config.Config{
		GRPCPort: "3200",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lg, err := logging.NewZapLogger(&config.Config{LogLevel: -1})
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.counterRepository = repositories.NewMockICounterRepository(ctrl)
			tt.fields.gaugeRepository = repositories.NewMockIGaugeRepository(ctrl)
			tt.prepare(&tt.fields)

			h := &IndexHandler{
				gaugeRepository:   tt.fields.gaugeRepository,
				counterRepository: tt.fields.counterRepository,
				lg:                lg,
			}

			RunTestServer(t, cfg, lg, nil)

			conn, err := grpc.NewClient(":"+cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatal(err)
			}
			defer conn.Close()

			client := metrics.NewMetricsServiceClient(conn)

		})
	}
}
