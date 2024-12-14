package models

import (
	"encoding/json"
	"log"
)

type Metric struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (m Metric) String() string {
	record, err := json.Marshal(m)
	if err != nil {
		log.Println("models/mitric: marshal err %w", err)
		return ""
	}

	return string(record)
}
