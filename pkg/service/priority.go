package service

import (
	"fmt"
)

type Priority string

const (
	PriorityHigh Priority = "high"
	PriorityLow  Priority = "low"
)

func (p *Priority) IsValid() error {
	if string(*p) == "" {
		return fmt.Errorf("Priority is required")
	}
	if *p != PriorityHigh && *p != PriorityLow {
		return fmt.Errorf("Invalid Priority: %v", *p)
	}
	return nil
}
