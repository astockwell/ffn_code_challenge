# FFN Challenge

## Build / Run
Can be built with `go build` (modules-enabled).

Can be downloaded and run from the latest [Release](https://github.com/astockwell/ffn_code_challenge/releases).

## Use
The server listens for HTTP on port :8080. The data is stored ephemerally in-memory, with Agents, Skills, and Priorities pre-seeded at runtime.

The following routes are supported:

- `/` - List of agents with the tasks currently assigned to them (if any). Example: `curl http://localhost:8080/`
- `/tasks/new` - Create a new task. Accepts a task object and will return that task, updated with the assigned agent if one was available. Example: `curl -X POST -d '{"priority":"high","required_skills":["skill1"]}' http://localhost:8080/tasks/new`
- `/tasks/complete` - Mark a task as completed. Example: `curl -X POST -d '{"id":2}' http://localhost:8080/tasks/complete`

## Testing
Tests can be run within ./cmd/agenttaskapi by running `go test`.

The following tests are defined and passing:
- Test_route_Tasks_New_POST
- Test_route_Tasks_New_POST/Simple_Assignment_goes_to_first_available_agent
- Test_route_Tasks_New_POST/Simple_Assignment_w/_existing_task_(1)_goes_to_next_available_agent
- Test_route_Tasks_New_POST/Simple_Assignment_w/_existing_tasks_(2)_goes_to_last_available_agent
- Test_route_Tasks_New_POST/Assignment_fails:_no_agent_w/_skills_available
- Test_route_Tasks_New_POST/Assignment_fails:_no_agent_available_for_priority
- Test_route_Tasks_New_POST/Assignment_of_higher_priority_proceeds_to_agent_w/_most_recently_assigned_task
- Test_route_Tasks_New_POST/Assignment_of_higher_priority_proceeds_to_agent_w/_most_recently_assigned_task,_excluding_other_busy_agents
- Test_route_Tasks_Update_Complete_POST
- Test_route_Tasks_Update_Complete_POST/Simple_task_completion
- Test_route_Tasks_Update_Complete_POST/Task_completion_with_>1_tasks_in_queue

## Questions / Answers
It seems that an agent can be assigned multiple active tasks, as long as priority is respected. Is that true?
- As long as the conditions are not violated and priority is respected, an agent can be assigned multiple tasks.

Will a task ever be assigned to more than one agent at a time?
- No

Do we need to know the history of a task's assignees?
- This is not a requirement in the exercise, but a design that allows this in the future would be ideal.

If an agent is working on a lower priority task and the system picks them for a higher priority assignment, what should happen to the lower priority task (e.g. should it be reassigned)?
- The lower-priority task should remained assigned to the agent.

Is there more than 1 completion state? That is to say, it seems that tasks can be "WIP" or "Complete", is there a "Cancelled" or other state?
- There only needs to be one completed state, but again, flexible design is appreciated.

Is it acceptible to use an in-memory, in-process (ephemeral) data store (which would be pre-seeded at runtime)?
- For the purpose of this exercise, an in-memory data store is fine.  A production system should be resilient to restarts, but there could be multiple ways to achieve that.


