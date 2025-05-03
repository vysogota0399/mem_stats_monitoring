package agent

import (
	"sync"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
)

type MetricsPool struct {
	pool     sync.Pool
	metricsL sync.Mutex
}

func NewMetricsPool() MetricsPool {
	return MetricsPool{
		pool: sync.Pool{
			New: func() any {
				return &models.Metric{}
			},
		},
		metricsL: sync.Mutex{},
	}
}

func (p *MetricsPool) Get() *models.Metric {
	return p.pool.Get().(*models.Metric)
}

func (p *MetricsPool) Put(m *models.Metric) {
	p.metricsL.Lock()
	defer p.metricsL.Unlock()

	p.pool.Put(m.Reset())
}

func (p *MetricsPool) Free(batch []*models.Metric) {
	for _, m := range batch {
		p.Put(m)
	}
}
