package agent

import (
	"sync"

	"github.com/vysogota0399/mem_stats_monitoring/internal/agent/models"
)

type MetricsPool struct {
	pool sync.Pool
}

func NewMetricsPool() MetricsPool {
	return MetricsPool{
		pool: sync.Pool{
			New: func() any {
				return &models.Metric{}
			},
		},
	}
}

func (p *MetricsPool) Get(name, mtype, value string) *models.Metric {
	m := p.pool.Get().(*models.Metric)
	m.Name = name
	m.Type = mtype
	m.Value = value
	return m
}

func (p *MetricsPool) Put(m *models.Metric) {
	m.Name = ""
	m.Type = ""
	m.Value = ""
	p.pool.Put(m)
}

func (p *MetricsPool) Free(batch []*models.Metric) {
	for _, m := range batch {
		p.Put(m)
	}
}
