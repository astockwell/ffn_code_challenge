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
