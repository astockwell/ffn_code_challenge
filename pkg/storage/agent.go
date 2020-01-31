package storage

import "sort"

import "fmt"

type Agent struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Skills Skills `json:"skills"`

	Tasks []*Task `json:"tasks,omitempty"`
}

type Agents []Agent

func BuildSeedAgents() []*Agent {
	return []*Agent{
		&Agent{
			Name:   "Adam",
			Skills: Skills{Skill1, Skill2},
			Tasks:  []*Task{},
		},
		&Agent{
			Name:   "Betty",
			Skills: Skills{Skill2, Skill3},
			Tasks:  []*Task{},
		},
		&Agent{
			Name:   "Charlie",
			Skills: Skills{Skill1},
			Tasks:  []*Task{},
		},
	}
}

// FilterForAvailableByPriority returns a slice of agents availabile for a task of the given priority
func (as *Agents) FilterForAvailableByPriority(targetPriority Priority) (availableAgents Agents, atLeastOneAgentAvailable bool) {
	availableAgents = []Agent{}

	for _, a := range *as {
		if !a.AvailableForAssignment(targetPriority) {
			continue
		}
		availableAgents = append(availableAgents, a)
	}

	return availableAgents, (len(availableAgents) > 0)
}

// FilterForNoTasksAssigned returns a slice of agents with no tasks assigned
func (as *Agents) FilterForNoTasksAssigned() (idleAgents Agents, atLeastOneAgentIdle bool) {
	idleAgents = []Agent{}

	for _, a := range *as {
		if len(a.Tasks) > 0 {
			continue
		}
		idleAgents = append(idleAgents, a)
	}

	return idleAgents, (len(idleAgents) > 0)
}

// SortByTaskCount sorts a slice of agents by the # of assigned tasks they have, lowest-to-highest
func (as *Agents) SortByTaskCount() {
	sort.Slice(*as, func(i, j int) bool {
		return len((*as)[i].Tasks) < len((*as)[j].Tasks)
	})
}

// SortByTaskStartTime sorts a slice of agents by the most-recently started task time
func (as *Agents) SortByTaskStartTime() error {
	_, atLeastOneAgentIdle := as.FilterForNoTasksAssigned()
	if atLeastOneAgentIdle {
		return fmt.Errorf("At least one agent is idle, cannot complete sort by task start")
	}
	sort.Slice(*as, func(i, j int) bool {
		return (*as)[i].Tasks[0].AssignmentTime.Unix() > (*as)[j].Tasks[0].AssignmentTime.Unix()
	})
	return nil
}

// PluckRandomAgent returns a random agent from the receiver
// TODO: NOT YET IMPLEMENTED
func (as *Agents) PluckRandomAgent() (Agent, error) {
	if len(*as) < 1 {
		return Agent{}, fmt.Errorf("No agents to pick from")
	}
	return (*as)[0], nil
}

func (a *Agent) HasSkills(ss Skills) bool {
	for _, skill := range ss {
		if !a.Skills.Includes(skill) {
			return false
		}
	}
	return true
}

func (a *Agent) AvailableForAssignment(p Priority) bool {
	for _, t := range a.Tasks {
		switch t.Priority {
		case PriorityHigh:
			// This is the highest priority, an agent with a task
			// of this priority is unavailable in all cases.
			return false
		case PriorityLow:
			if p == PriorityLow {
				return false
			}
		}
	}
	return true
}

func (a *Agent) Clone() Agent {
	return Agent{
		ID:     a.ID,
		Name:   a.Name,
		Skills: a.Skills,
		Tasks:  a.Tasks,
	}
}

func (a *Agent) SlimClone() Agent {
	return Agent{
		ID:     a.ID,
		Name:   a.Name,
		Skills: a.Skills,
	}
}
