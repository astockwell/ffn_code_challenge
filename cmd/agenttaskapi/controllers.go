package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astockwell/ffn/pkg/service"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// route_Index lists all agents and their tasks
func route_Index(dso *DataSourceOrchestration) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, rp httprouter.Params) {
		log.Tracef("route_Index(): Started")

		agents, err := dso.Store.ListAgents()
		if err != nil {
			log.Errorf("route_Index() --> Retrieving Agents from Store: %v", err)
			dso.Renderer.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Error retrieving agents list from data store"})
			return
		}

		dso.Renderer.JSON(w, http.StatusOK, agents)
	}
}

// route_Tasks_New_POST assigns a task to an Agent, if available and permissable
func route_Tasks_New_POST(dso *DataSourceOrchestration) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, rp httprouter.Params) {
		log.Tracef("route_Tasks_New_POST(): Started")

		// Parse request body JSON
		var newTask service.Task
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&newTask)
		if err != nil {
			log.Warnf("route_Tasks_New_POST() --> json.Decode(&newTask): %v", err)
			dso.Renderer.JSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("JSON decode of request body failed: %v", err)})
			return
		}
		log.Tracef("route_Tasks_New_POST(): Decoded JSON to task: %v", newTask)

		// Validate task (this could be omitted, but we can present a nicer error message to the end user this way)
		err = newTask.IsValid()
		if err != nil {
			log.Warnf("route_Tasks_New_POST() --> !newTask.IsValid(): %v; Task: %#v", err, newTask)
			dso.Renderer.JSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("New Task is invalid: %v", err)})
			return
		}
		log.Tracef("route_Tasks_New_POST(): newTask is valid")

		// Assign task
		agentAssignedID, taskID, err := dso.Store.AddTaskToAgent(&newTask)
		if err != nil {
			log.Warnf("route_Tasks_New_POST() --> Store.AddTaskToAgent(newTask): %v; Task: %#v", err, newTask)
			dso.Renderer.JSON(w, http.StatusConflict, map[string]string{"error": fmt.Sprintf("Could not assign task: %v", err)})
			return
		}
		log.Tracef("route_Tasks_New_POST(): newTask (ID: %v) assigned to agent (ID: %v) successfully", taskID, agentAssignedID)

		// Fetch assigned task details for response
		assignedTask, err := dso.Store.FindTaskWithAgent(taskID)
		if err != nil {
			log.Errorf("route_Tasks_New_POST() --> Store.FindTask(taskID): %v", err)
			dso.Renderer.JSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Could not find saved task (%v) in data store: %v", taskID, err)})
			return
		}
		log.Tracef("route_Tasks_New_POST(): assignedTask fetched")

		dso.Renderer.JSON(w, http.StatusCreated, assignedTask)
	}
}

// route_Tasks_Update_Complete_POST marks the task as complete via the given task ID
func route_Tasks_Update_Complete_POST(dso *DataSourceOrchestration) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, rp httprouter.Params) {
		log.Tracef("route_Tasks_Update_Complete_POST(): Started")

		// Parse request body JSON
		var task service.Task
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&task)
		if err != nil {
			log.Warnf("route_Tasks_New_POST() --> json.Decode(&task): %v", err)
			dso.Renderer.JSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("JSON decode of request body failed: %v", err)})
			return
		}
		log.Tracef("route_Tasks_Update_Complete_POST(): Decoded JSON to task: %v", task)

		err = dso.Store.MarkAsCompleted(task.ID)
		if err != nil {
			log.Warnf("route_Tasks_New_POST() --> Store.MarkAsCompleted(task.ID): %v", err)
			dso.Renderer.JSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("Error occurred marking task as completed: %v", err)})
			return
		}

		dso.Renderer.JSON(w, http.StatusOK, nil)
	}
}
