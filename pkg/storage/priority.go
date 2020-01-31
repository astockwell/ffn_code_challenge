package storage

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

// // UnmarshalJSON performs validations at the time of Unmarshalling JSON
// func (p *Priority) UnmarshalJSON(b []byte) error {
// 	log.Tracef("Priority.UnmarshalJSON(): Started")

// 	if len(b) < 1 {
// 		return fmt.Errorf("priority is required")
// 	}
// 	var priority string
// 	err := json.Unmarshal(b, &priority)
// 	if err != nil {
// 		return errors.Wrap(err, "json.Unmarshal()")
// 	}

// 	switch priority {
// 	case string(PriorityHigh):
// 		*p = PriorityHigh
// 	case string(PriorityLow):
// 		*p = PriorityLow
// 	default:
// 		return fmt.Errorf("Invalid Priority: %v", priority)
// 	}

// 	return nil
// }
