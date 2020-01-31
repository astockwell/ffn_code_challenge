package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/astockwell/ffn/pkg/service"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/unrolled/render"
)

func buildSeededTestStore(t *testing.T) *service.Store {
	// Setup data store
	store := &service.Store{}

	// Seed data store
	agents := service.BuildSeedAgents()
	err := store.AddAgents(agents)
	if err != nil {
		t.Fatal(err)
	}

	return store
}

func Test_route_Tasks_New_POST(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	// log.SetFormatter(&prefixed.TextFormatter{FullTimestamp: true, TimestampFormat: "2006/01/02 15:04:05", ForceFormatting: true})
	// log.SetLevel(log.TraceLevel)

	type testRequest struct {
		Priority  string   `json:"priority,omitempty"`
		ReqSkills []string `json:"required_skills,omitempty"`
	}
	type testResponseAgent struct {
		ID     uint     `json:"id"`
		Name   string   `json:"name"`
		Skills []string `json:"skills"`
	}
	type testResponse struct {
		ID             uint              `json:"id"`
		Priority       string            `json:"priority"`
		RequiredSkills []string          `json:"required_skills"`
		AssignedAgent  testResponseAgent `json:"assigned_agent"`
		TaskState      int               `json:"task_state"`
	}

	tests := []struct {
		name                 string         // Test name
		store                *service.Store // Initial state of the data store prior to HTTP request
		postBody             testRequest    // HTTP request body (in struct form)
		wantStatus           int            // Expected HTTP response code
		wantResponseContains []string       // For validating errors
		wantResponse         *testResponse  // For validating successes
		wantStore            *service.Store // Expected state of the data store after HTTP request is complete
	}{
		{
			name: "Simple Assignment goes to first available agent",
			store: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{}},
			}, nil),
			postBody:   testRequest{Priority: "high", ReqSkills: []string{"skill1"}},
			wantStatus: http.StatusCreated,
			wantResponse: &testResponse{
				ID:             1,
				Priority:       "high",
				RequiredSkills: []string{"skill1"},
				AssignedAgent:  testResponseAgent{ID: 1, Name: "Adam", Skills: []string{"skill1", "skill2"}},
				TaskState:      0,
			},
			wantStore: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{}},
			}, nil),
		},
		{
			name: "Simple Assignment w/ existing task (1) goes to next available agent",
			store: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{}},
			}, nil),
			postBody:   testRequest{Priority: "high", ReqSkills: []string{"skill1"}},
			wantStatus: http.StatusCreated,
			wantResponse: &testResponse{
				ID:             2,
				Priority:       "high",
				RequiredSkills: []string{"skill1"},
				AssignedAgent:  testResponseAgent{ID: 3, Name: "Charlie", Skills: []string{"skill1"}},
				TaskState:      0,
			},
			wantStore: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{
					&service.Task{ID: 2, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
			}, nil),
		},
		{
			name: "Simple Assignment w/ existing tasks (2) goes to last available agent",
			store: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{
					&service.Task{ID: 2, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
			}, nil),
			postBody:   testRequest{Priority: "high", ReqSkills: []string{"skill3"}},
			wantStatus: http.StatusCreated,
			wantResponse: &testResponse{
				ID:             3,
				Priority:       "high",
				RequiredSkills: []string{"skill3"},
				AssignedAgent:  testResponseAgent{ID: 2, Name: "Betty", Skills: []string{"skill2", "skill3"}},
				TaskState:      0,
			},
			wantStore: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{
					&service.Task{ID: 3, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill3}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{
					&service.Task{ID: 2, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
			}, nil),
		},
		{
			name:                 "Assignment fails: no agent w/ skills available",
			store:                service.NewStore([]*service.Agent{}, nil),
			postBody:             testRequest{Priority: "high", ReqSkills: []string{"skill1"}},
			wantStatus:           http.StatusConflict,
			wantResponseContains: []string{`{"error":"Could not assign task: No existing agents possess the required skills for this task"}`},
			wantStore:            service.NewStore([]*service.Agent{}, nil),
		},
		{
			name: "Assignment fails: no agent available for priority",
			store: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{
					&service.Task{ID: 2, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
			}, nil),
			postBody:             testRequest{Priority: "high", ReqSkills: []string{"skill1"}},
			wantStatus:           http.StatusConflict,
			wantResponseContains: []string{`{"error":"Could not assign task: No agents are currently available for this task priority"}`},
			wantStore: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{
					&service.Task{ID: 2, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
			}, nil),
		},
		{
			name: "Assignment of higher priority proceeds to agent w/ most recently assigned task",
			store: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityLow, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP, AssignmentTime: time.Now().Add(-2 * time.Hour)},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{
					&service.Task{ID: 2, Priority: service.PriorityLow, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP, AssignmentTime: time.Now().Add(-1 * time.Hour)},
				}},
			}, nil),
			postBody:   testRequest{Priority: "high", ReqSkills: []string{"skill1"}},
			wantStatus: http.StatusCreated,
			wantResponse: &testResponse{
				ID:             3,
				Priority:       "high",
				RequiredSkills: []string{"skill1"},
				AssignedAgent:  testResponseAgent{ID: 3, Name: "Charlie", Skills: []string{"skill1"}},
				TaskState:      0,
			},
			wantStore: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityLow, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{
					&service.Task{ID: 3, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
					&service.Task{ID: 2, Priority: service.PriorityLow, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
			}, nil),
		},
		{
			name: "Assignment of higher priority proceeds to agent w/ most recently assigned task, excluding other busy agents",
			store: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityLow, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP, AssignmentTime: time.Now().Add(-2 * time.Hour)},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{
					&service.Task{ID: 3, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
					&service.Task{ID: 2, Priority: service.PriorityLow, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
			}, nil),
			postBody:   testRequest{Priority: "high", ReqSkills: []string{"skill1"}},
			wantStatus: http.StatusCreated,
			wantResponse: &testResponse{
				ID:             4,
				Priority:       "high",
				RequiredSkills: []string{"skill1"},
				AssignedAgent:  testResponseAgent{ID: 1, Name: "Adam", Skills: []string{"skill1", "skill2"}},
				TaskState:      0,
			},
			wantStore: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 4, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
					&service.Task{ID: 1, Priority: service.PriorityLow, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{
					&service.Task{ID: 3, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
					&service.Task{ID: 2, Priority: service.PriorityLow, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
			}, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare web server components
			renderer := render.New()
			dso := &DataSourceOrchestration{
				Renderer: renderer,
				Store:    tt.store,
			}

			// Marshal test body to JSON
			reqJSON, err := json.Marshal(tt.postBody)
			if err != nil {
				t.Fatal(err)
			}

			// Build test request
			r, err := http.NewRequest("POST", "/tasks/new", bytes.NewReader(reqJSON))
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Add("Content-Type", "application/json; charset=UTF-8")

			// Build response recorder
			w := httptest.NewRecorder()

			// Execute test request
			route_Tasks_New_POST(dso)(w, r, httprouter.Params{})
			// spew.Dump(w))
			// fmt.Println(w.Body)

			// Reset timestamps from store tasks
			tt.store.TESTING_resetTaskAssignmentTimes()

			// Assertions
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantStore, tt.store)

			// For errors, validate the error message(s) returned in the HTTP response
			if len(tt.wantResponseContains) > 0 {
				body := w.Body.String()
				for _, contents := range tt.wantResponseContains {
					if !strings.Contains(body, contents) {
						t.Errorf("Want response body to contain '%v'\nResponse Body was: %v", contents, body)
					}
				}
			}

			// For successes, validate the Task{} information returned in the HTTP response
			if tt.wantResponse != nil {
				wantRespJSON, err := json.Marshal(tt.wantResponse) // Marshal expected test result struct --> JSON
				if err != nil {
					t.Fatal(err)
				}
				var wantTask service.Task
				err = json.Unmarshal(wantRespJSON, &wantTask) // Unmarshal expected test result JSON --> Task{}
				if err != nil {
					t.Fatal(err)
				}
				var gotTask service.Task
				err = json.Unmarshal(w.Body.Bytes(), &gotTask) // Unmarshal POST HTTP response body --> Task{}
				if err != nil {
					t.Fatal(err)
				}
				gotTask.AssignmentTime = time.Time{} // Clear HTTP response Task{} assignment timestamp

				assert.Equal(t, wantTask, gotTask)
			}

		})
	}
}

func Test_route_Tasks_Update_Complete_POST(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	// log.SetFormatter(&prefixed.TextFormatter{FullTimestamp: true, TimestampFormat: "2006/01/02 15:04:05", ForceFormatting: true})
	// log.SetLevel(log.TraceLevel)

	type testRequest struct {
		ID uint `json:"id"`
	}

	tests := []struct {
		name       string         // Test name
		store      *service.Store // Initial state of the data store prior to HTTP request
		postBody   testRequest    // HTTP request body (in struct form)
		wantStatus int            // Expected HTTP response code
		wantStore  *service.Store // Expected state of the data store after HTTP request is complete
	}{
		{
			name: "Simple task completion",
			store: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{}},
			}, nil),
			postBody:   testRequest{ID: 1},
			wantStatus: http.StatusOK,
			wantStore: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{}},
			}, []*service.Task{
				&service.Task{ID: 1, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskComplete, AssignedAgent: &service.Agent{ID: 1, Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}}},
			}),
		},
		{
			name: "Task completion with >1 tasks in queue",
			store: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 1, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
					&service.Task{ID: 2, Priority: service.PriorityLow, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{}},
			}, nil),
			postBody:   testRequest{ID: 1},
			wantStatus: http.StatusOK,
			wantStore: service.NewStore([]*service.Agent{
				&service.Agent{Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}, Tasks: []*service.Task{
					&service.Task{ID: 2, Priority: service.PriorityLow, ReqSkills: service.Skills{service.Skill1}, State: service.TaskInWIP},
				}},
				&service.Agent{Name: "Betty", Skills: service.Skills{service.Skill2, service.Skill3}, Tasks: []*service.Task{}},
				&service.Agent{Name: "Charlie", Skills: service.Skills{service.Skill1}, Tasks: []*service.Task{}},
			}, []*service.Task{
				&service.Task{ID: 1, Priority: service.PriorityHigh, ReqSkills: service.Skills{service.Skill1}, State: service.TaskComplete, AssignedAgent: &service.Agent{ID: 1, Name: "Adam", Skills: service.Skills{service.Skill1, service.Skill2}}},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare web server components
			renderer := render.New()
			dso := &DataSourceOrchestration{
				Renderer: renderer,
				Store:    tt.store,
			}

			// Marshal test body to JSON
			reqJSON, err := json.Marshal(tt.postBody)
			if err != nil {
				t.Fatal(err)
			}

			// Build test request
			r, err := http.NewRequest("POST", "/tasks/complete", bytes.NewReader(reqJSON))
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Add("Content-Type", "application/json; charset=UTF-8")

			// Build response recorder
			w := httptest.NewRecorder()

			// Execute test request
			route_Tasks_Update_Complete_POST(dso)(w, r, httprouter.Params{})
			// spew.Dump(w))
			// fmt.Println(w.Body)

			// Reset timestamps from store tasks
			tt.store.TESTING_resetTaskAssignmentTimes()

			// Assertions
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantStore, tt.store)

		})
	}
}
