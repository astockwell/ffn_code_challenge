package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Store (memory) keeps data in memory
type Store struct {
	sync.RWMutex
	agents         []*Agent
	completedTasks []*Task
}

func NewStore(agents []*Agent, completed []*Task) *Store {
	for i := 0; i < len(agents); i++ {
		agents[i].ID = uint(i + 1)
	}
	return &Store{
		agents:         agents,
		completedTasks: completed,
	}
}

func (s *Store) AddAgents(agents []*Agent) error {
	for i := 0; i < len(agents); i++ {
		agents[i].ID = s.NextAgentID()
		s.Lock()
		s.agents = append(s.agents, agents[i])
		s.Unlock()
	}

	return nil
}

func (s *Store) FindAgent(agentID uint) (*Agent, error) {
	s.RLock()
	defer s.RUnlock()

	for _, a := range s.agents {
		if a.ID == agentID {
			return a, nil
		}
	}

	return nil, fmt.Errorf("Agent not found")
}

func (s *Store) ListAgents() ([]*Agent, error) {
	s.RLock()
	defer s.RUnlock()

	// Build + return copies
	as := make([]*Agent, 0, len(s.agents))
	for _, a := range s.agents {
		as = append(as, a)
	}

	return as, nil
}

func (s *Store) FindTask(taskID uint) (*Task, error) {
	s.RLock()
	defer s.RUnlock()

	for _, a := range s.agents {
		for _, t := range a.Tasks {
			if t.ID == taskID {
				return t, nil
			}
		}
	}

	return nil, fmt.Errorf("Task not found")
}

func (s *Store) DeleteTask(taskID uint) error {
	s.Lock()
	defer s.Unlock()

	for i := 0; i < len(s.agents); i++ {
		for j := 0; j < len(s.agents[i].Tasks); j++ {
			if s.agents[i].Tasks[j].ID == taskID {
				// Delete by snipping task from slice
				s.agents[i].Tasks = append(s.agents[i].Tasks[:j], s.agents[i].Tasks[j+1:]...)
				return nil
			}
		}
	}

	return fmt.Errorf("Task not found")
}

func (s *Store) FindTaskWithAgent(taskID uint) (Task, error) {
	s.RLock()
	defer s.RUnlock()

	for _, a := range s.agents {
		for _, t := range a.Tasks {
			if t.ID == taskID {
				result := t.Clone()
				agent := a.SlimClone()
				result.AssignedAgent = &agent
				return result, nil
			}
		}
	}

	return Task{}, fmt.Errorf("Task not found")
}

// AddTaskToAgent adds a task to an agents list, or returns an error if no agents are available
func (s *Store) AddTaskToAgent(t *Task) (assignedAgentID uint, taskID uint, err error) {
	// Ensure task is valid
	err = t.IsValid()
	if err != nil {
		return 0, 0, errors.Wrap(err, "task.IsValid()")
	}

	// Find agents with task required skills
	skilledAgentPool, ok := s.FindAgentsWithNecessarySkills(t.ReqSkills)
	if !ok {
		return 0, 0, fmt.Errorf("No existing agents possess the required skills for this task")
	}

	// Filter agents for availability for task priority
	availableAgentPool, ok := skilledAgentPool.FilterForAvailableByPriority(t.Priority)
	if !ok {
		return 0, 0, fmt.Errorf("No agents are currently available for this task priority")
	}

	// Prefer agents with no tasks assigned
	var selectedAgent Agent
	idleAgents, ok := availableAgentPool.FilterForNoTasksAssigned()
	if ok {
		// Assign to random agent; ignore errors as we know len(idleAgents) > 0
		selectedAgent, _ = idleAgents.PluckRandomAgent()

		err = s.addTaskToAgentPush(selectedAgent.ID, t)
		if err != nil {
			return 0, 0, errors.Wrap(err, "s.addTaskToAgentPush()")
		}
		return selectedAgent.ID, t.ID, nil
	}

	// Order available agents by current task start time;
	// ignore errors because we know all agents in availableAgentPool have tasks
	_ = availableAgentPool.SortByTaskStartTime()
	selectedAgent = availableAgentPool[0]

	err = s.addTaskToAgentUnshift(selectedAgent.ID, t)
	if err != nil {
		return 0, 0, errors.Wrap(err, "s.addTaskToAgentUnshift()")
	}
	return selectedAgent.ID, t.ID, nil
}

// addTaskToAgentUnshift adds task to front of an agent's queue, effectively assigning it to them
func (s *Store) addTaskToAgentUnshift(agentID uint, t *Task) error {
	agent, err := s.FindAgent(agentID)
	if err != nil {
		return errors.Wrap(err, "s.FindAgent(agentID)")
	}

	s.Lock()
	defer s.Unlock()

	t.ID = s.NextTaskID()
	t.AssignmentTime = time.Now()
	t.State = TaskInWIP
	agent.Tasks = append([]*Task{t}, agent.Tasks...)
	return nil
}

// addTaskToAgentPush adds task to back of an agent's queue
func (s *Store) addTaskToAgentPush(agentID uint, t *Task) error {
	agent, err := s.FindAgent(agentID)
	if err != nil {
		return errors.Wrap(err, "s.FindAgent(agentID)")
	}

	s.Lock()
	defer s.Unlock()

	t.ID = s.NextTaskID()
	t.AssignmentTime = time.Now()
	t.State = TaskInWIP
	agent.Tasks = append(agent.Tasks, t)
	return nil
}

func (s *Store) MarkAsCompleted(taskID uint) error {
	task, err := s.FindTaskWithAgent(taskID)
	if err != nil {
		return errors.Wrap(err, "s.FindTaskWithAgent(taskID)")
	}

	s.Lock()

	// Flag as complete
	task.State = TaskComplete

	// Add to completed list
	s.completedTasks = append(s.completedTasks, &task)

	s.Unlock()

	// Purge from Agent's assignments
	err = s.DeleteTask(taskID)
	if err != nil {
		return errors.Wrap(err, "s.DeleteTask(taskID)")
	}

	return nil
}

// NextAgentID returns the next available ID that should be used for a new Agent{}
func (s *Store) NextAgentID() uint {
	id := uint(1)
	for _, agent := range s.agents {
		if agent.ID >= id {
			id = agent.ID + 1
		}
	}
	return id
}

// NextTaskID returns the next available ID that should be used for a new Task{}
func (s *Store) NextTaskID() uint {
	id := uint(1)
	for _, agent := range s.agents {
		for _, task := range agent.Tasks {
			if task.ID >= id {
				id = task.ID + 1
			}
		}
	}
	return id
}

// FindAgentsWithNecessarySkills returns a slice of agents with task required skills
func (s *Store) FindAgentsWithNecessarySkills(ss Skills) (skilledAgents Agents, atLeastOneFound bool) {
	s.RLock()
	defer s.RUnlock()

	skillMatchedAgents := Agents{}
	for _, agent := range s.agents {
		if !agent.HasSkills(ss) {
			continue
		}
		skillMatchedAgents = append(skillMatchedAgents, *agent)
	}

	return skillMatchedAgents, (len(skillMatchedAgents) > 0)
}

// TESTING_resetTaskAssignmentTimes is for testing purposes; resets all Task.AssignmentTime values to time.Time{}
func (s *Store) TESTING_resetTaskAssignmentTimes() {
	s.Lock()
	defer s.Unlock()

	for i := 0; i < len(s.agents); i++ {
		for j := 0; j < len(s.agents[i].Tasks); j++ {
			s.agents[i].Tasks[j].AssignmentTime = time.Time{}
		}
	}
}
