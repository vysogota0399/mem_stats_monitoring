package agent

import (
	"sync"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
)

type MetricsPool struct {
	pool sync.Pool
}

func NewMetricsPool() *MetricsPool {
	return &MetricsPool{
		pool: sync.Pool{
			New: func() any {
				return &models.Metric{}
			},
		},
	}
}

func (p *MetricsPool) Get() *models.Metric {
	return p.pool.Get().(*models.Metric)
}

func (p *MetricsPool) Put(m *models.Metric) {
	p.pool.Put(m.Reset())
}
