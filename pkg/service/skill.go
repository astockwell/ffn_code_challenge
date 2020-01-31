package service

import (
	"fmt"
)

type Skill string

type Skills []Skill

const (
	Skill1 Skill = "skill1"
	Skill2 Skill = "skill2"
	Skill3 Skill = "skill3"
)

func (s *Skills) IsValid() error {
	if len(*s) < 1 {
		return fmt.Errorf("At least one Required Skill is required")
	}
	for _, skill := range *s {
		if skill != Skill1 && skill != Skill2 && skill != Skill3 {
			return fmt.Errorf("Invalid Skill: %v", skill)
		}
	}
	return nil
}

func (s *Skills) Includes(targetSkill Skill) bool {
	for _, skill := range *s {
		if skill == targetSkill {
			return true
		}
	}
	return false
}

// // UnmarshalJSON performs validations at the time of Unmarshalling JSON
// func (s Skills) UnmarshalJSON(b []byte) error {
// 	var skills []string
// 	err := json.Unmarshal(b, &skills)
// 	if err != nil {
// 		return errors.Wrap(err, "json.Unmarshal")
// 	}

// 	for _, skill := range skills {
// 		switch skill {
// 		case string(Skill1):
// 			s = append(s, Skill1)
// 		case string(Skill2):
// 			s = append(s, Skill2)
// 		case string(Skill3):
// 			s = append(s, Skill3)
// 		default:
// 			return fmt.Errorf("Invalid Skill: %v", skill)
// 		}
// 	}

// 	return nil
// }
