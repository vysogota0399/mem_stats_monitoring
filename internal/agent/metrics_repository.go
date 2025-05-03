package agent

import (
	"context"
	"sync"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/storage"
)

type MetricsRepository struct {
	mu      sync.RWMutex
	storage *storage.Memory
	pool    MetricsPool
}

func NewMetricsRepository(st *storage.Memory) *MetricsRepository {
	return &MetricsRepository{
		mu:      sync.RWMutex{},
		storage: st,
		pool:    NewMetricsPool(),
	}
}

func (r *MetricsRepository) Get(name, mtype string) (*models.Metric, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m := r.pool.Get(name, mtype, "")

	err := r.storage.Get(m)
	if err != nil {
		r.pool.Put(m)
		return nil, err
	}

	return m, nil
}

func (r *MetricsRepository) New(name, mtype, value string) *models.Metric {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.pool.Get(name, mtype, value)
}

func (r *MetricsRepository) Release(metrics ...*models.Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pool.Free(metrics)
}

func (r *MetricsRepository) SaveAndRelease(ctx context.Context, m *models.Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	defer r.pool.Put(m)

	return r.storage.Set(ctx, m)
}

func (r *MetricsRepository) SafeRead(m *models.Metric) (name, mtype, value string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	name = m.Name
	mtype = m.Type
	value = m.Value
	return
}
