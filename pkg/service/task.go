package service

import "time"

type Task struct {
	ID             uint      `json:"id"`
	Priority       Priority  `json:"priority"`
	ReqSkills      Skills    `json:"required_skills"`
	AssignedAgent  *Agent    `json:"assigned_agent,omitempty"`
	AssignmentTime time.Time `json:"assignment_time"`
	State          TaskState `json:"task_state"`
}

func (t *Task) IsValid() error {
	if err := t.Priority.IsValid(); err != nil {
		return err
	}
	if err := t.ReqSkills.IsValid(); err != nil {
		return err
	}
	return nil
}

func (t *Task) Clone() Task {
	return Task{
		ID:             t.ID,
		Priority:       t.Priority,
		ReqSkills:      t.ReqSkills,
		AssignedAgent:  t.AssignedAgent,
		AssignmentTime: t.AssignmentTime,
		State:          t.State,
	}
}
