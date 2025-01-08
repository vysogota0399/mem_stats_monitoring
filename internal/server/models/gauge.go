package models

import (
	"fmt"
)

const GaugeType = "gauge"

type Gauge struct {
	Value float64 `json:"value"`
	Name  string  `json:"name"`
}

func (g Gauge) StringValue() string {
	return fmt.Sprintf("%v", g.Value)
}
