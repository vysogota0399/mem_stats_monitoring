package agent

import (
	"context"
	"sync"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
)

type MetricsRepository struct {
	mu      sync.Mutex
	storage *storage.Memory
	pool    MetricsPool
}

func NewMetricsRepository(st *storage.Memory) *MetricsRepository {
	return &MetricsRepository{
		mu:      sync.Mutex{},
		storage: st,
		pool:    NewMetricsPool(),
	}
}

func (r *MetricsRepository) Get(name, mtype string) (*models.Metric, error) {
	m := r.pool.Get()
	m.Type = mtype
	m.Name = name
	err := r.storage.Get(m)
	if err != nil {
		r.pool.Put(m)
		return nil, err
	}

	return m, nil
}

func (r *MetricsRepository) New(name, mtype, value string) *models.Metric {
	m := r.pool.Get()
	m.Name = name
	m.Type = mtype
	m.Value = value
	return m
}

func (r *MetricsRepository) Release(metrics ...*models.Metric) {
	r.pool.Free(metrics)
}

func (r *MetricsRepository) SaveAndRelease(ctx context.Context, m *models.Metric) error {
	defer r.pool.Put(m)

	return r.storage.Set(ctx, m)
}
