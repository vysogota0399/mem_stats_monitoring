package models

type Gauge struct {
	Value float64 `json:"value"`
	Name  string  `json:"name"`
}
