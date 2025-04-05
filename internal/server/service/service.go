package service

import (
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

// Service выступает в качестве контейнера для хранения других сервисов отвечающих за бизнес логику.
type Service struct {
	UpdateMetricService  *UpdateMetricService
	UpdateMetricsService *UpdateMetricsService
}

func New(s storage.Storage) *Service {
	return &Service{
		UpdateMetricService: &UpdateMetricService{
			gaugeRep:   repositories.NewGauge(s),
			counterRep: repositories.NewCounter(s),
		},
		UpdateMetricsService: &UpdateMetricsService{
			gaugeRep:   repositories.NewGauge(s),
			counterRep: repositories.NewCounter(s),
		},
	}
}
