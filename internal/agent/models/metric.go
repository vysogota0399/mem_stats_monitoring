package models

import (
	"github.com/mailru/easyjson"
)

const GaugeType = "gauge"
const CounterType = "counter"

type Metric struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (m *Metric) Reset() *Metric {
	m.Name = ""
	m.Type = ""
	m.Value = ""
	return m
}

func (m Metric) String() string {
	record, err := easyjson.Marshal(m)
	if err != nil {
		return ""
	}

	return string(record)
}
