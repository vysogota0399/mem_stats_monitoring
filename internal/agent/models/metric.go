package models

import "encoding/json"

type Metric struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (m Metric) String() string {
	record, err := json.Marshal(m)
	if err != nil {
		return err.Error()
	}

	return string(record)
}
