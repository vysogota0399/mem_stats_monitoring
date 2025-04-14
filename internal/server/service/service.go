package service

import (
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
	"github.com/vysogota0399/mem_stats_monitoring/internal/utils/logging"
)

// Service выступает в качестве контейнера для хранения других сервисов отвечающих за бизнес логику.
type Service struct {
	UpdateMetricService  *UpdateMetricService
	UpdateMetricsService *UpdateMetricsService
}

func New(s storage.Storage, lg *logging.ZapLogger) *Service {
	return &Service{
		UpdateMetricService: &UpdateMetricService{
			gaugeRep:   repositories.NewGauge(s, lg),
			counterRep: repositories.NewCounter(s, lg),
		},
		UpdateMetricsService: &UpdateMetricsService{
			gaugeRep:   repositories.NewGauge(s, lg),
			counterRep: repositories.NewCounter(s, lg),
		},
	}
}
