package service

import (
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/repositories"
	"github.com/vysogota0399/mem_stats_monitoring/internal/server/storage"
)

type Service struct {
	UpdateMetricService UpdateMetricService
}

func New(s storage.Storage) *Service {
	return &Service{
		UpdateMetricService: UpdateMetricService{
			gaugeRep:   repositories.NewGauge(s),
			counterRep: repositories.NewCounter(s),
		},
	}
}
