package grpc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/config"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/entities"
	"github.com/vysogota0399/mem_stats_monitoring/pkg/gen/services/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Reporter struct {
	client metrics.MetricsServiceClient
	lg     *logging.ZapLogger
	rep    *agent.MetricsRepository
}

func NewReporter(ctx context.Context, cfg *config.Config, rep *agent.MetricsRepository, lg *logging.ZapLogger) (*Reporter, error) {
	conn, err := grpc.NewClient(":"+cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server %s: %w", cfg.GRPCPort, err)
	}

	go func() {
		<-ctx.Done()
		lg.InfoCtx(ctx, "GC grpc client")
		if err := conn.Close(); err != nil {
			lg.ErrorCtx(ctx, "failed to close grpc client", zap.Error(err))
		}
	}()

	return &Reporter{
		client: metrics.NewMetricsServiceClient(conn),
		lg:     lg,
		rep:    rep,
	}, nil
}

func (r *Reporter) UpdateMetric(ctx context.Context, mType, mName, value string) error {
	item, err := metricToItem(mType, mName, value)
	if err != nil {
		return fmt.Errorf("failed to convert metric to item: %w", err)
	}

	in := &metrics.UpdateMetricParams{
		Item: item,
	}

	_, err = r.client.Update(ctx, in)
	if err != nil {
		return fmt.Errorf("failed to update metric: %w", err)
	}

	return nil
}

func (r *Reporter) UpdateMetrics(ctx context.Context, data []*models.Metric) error {
	metricsBody := make([]*metrics.Item, 0, len(data))
	for _, m := range data {
		name, mType, value := r.rep.SafeRead(m)
		item, err := metricToItem(mType, name, value)
		if err != nil {
			return fmt.Errorf("failed to convert metric to item: %w", err)
		}

		metricsBody = append(metricsBody, item)
	}

	in := &metrics.UpdateMetricsBatchParams{
		Item: metricsBody,
	}

	_, err := r.client.UpdateBatch(ctx, in)
	if err != nil {
		return fmt.Errorf("failed to update metrics: %w", err)
	}

	return nil
}

func metricToItem(mType, mName, value string) (*metrics.Item, error) {
	var item *metrics.Item
	switch mType {
	case models.GaugeType:
		gauge, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse gauge: %w", err)
		}

		item = &metrics.Item{
			Metric: &metrics.Item_Gauge{
				Gauge: &entities.Gauge{
					Value: gauge,
					Name:  mName,
				},
			},
		}
	case models.CounterType:
		counter, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse counter: %w", err)
		}

		item = &metrics.Item{
			Metric: &metrics.Item_Counter{
				Counter: &entities.Counter{
					Value: counter,
					Name:  mName,
				},
			},
		}
	default:
		return nil, fmt.Errorf("unknown metric type: %s", mType)
	}

	return item, nil
}
