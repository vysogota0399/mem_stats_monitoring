package models

import (
	"fmt"
)

const CounterType = "counter"

type Counter struct {
	ID    int64  `json:"id,omitempty"`
	Value int64  `json:"value"`
	Name  string `json:"name"`
}

func (c Counter) StringValue() string {
	return fmt.Sprintf("%d", c.Value)
}
