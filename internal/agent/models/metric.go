package models

import (
	"encoding/json"
)

const GaugeType = "gauge"
const CounterType = "counter"

type Metric struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (m Metric) String() string {
	record, err := json.Marshal(m)
	if err != nil {
		return ""
	}

	return string(record)
}
