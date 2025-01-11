package models

import (
	"fmt"
)

const GaugeType = "gauge"

type Gauge struct {
	ID    int64   `json:"id,omitempty"`
	Value float64 `json:"value"`
	Name  string  `json:"name"`
}

func (g Gauge) StringValue() string {
	return fmt.Sprintf("%v", g.Value)
}
