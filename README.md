# Notes
Task{
	ID
	Priority
	ReqSkills []string
	AssignedAgent *Agent
	AssignmentTime *time.Time
	State?
}
Agent{
	ID
	Skills []string
}

# Questions / Answers
It seems that an agent can be assigned multiple active tasks, as long as priority is respected. Is that true?
	-> As long as the conditions are not violated and priority is respected, an agent can be assigned multiple tasks.
Will a task ever be assigned to more than one agent at a time?
	-> No
Do we need to know the history of a task's assignees?
	-> This is not a requirement in the exercise, but a design that allows this in the future would be ideal.
If an agent is working on a lower priority task and the system picks them for a higher priority assignment, what should happen to the lower priority task (e.g. should it be reassigned)?
	-> The lower-priority task should remained assigned to the agent.
Is there more than 1 completion state? That is to say, it seems that tasks can be "WIP" or "Complete", is there a "Cancelled" or other state?
	-> There only needs to be one completed state, but again, flexible design is appreciated.
Is it acceptible to use an in-memory, in-process (ephemeral) data store (which would be pre-seeded at runtime)?
	-> For the purpose of this exercise, an in-memory data store is fine.  A production system should be resilient to restarts, but there could be multiple ways to achieve that.

