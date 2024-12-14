package models

import "fmt"

type Counter struct {
	Value int64  `json:"value"`
	Name  string `json:"name"`
}

func (c Counter) StringValue() string {
	return fmt.Sprintf("%d", c.Value)
}
