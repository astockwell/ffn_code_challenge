package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/astockwell/ffn/pkg/storage"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/unrolled/render"
)

func buildSeededTestStore(t *testing.T) *storage.Store {
	// Setup data store
	store := &storage.Store{}

	// Seed data store
	agents := storage.BuildSeedAgents()
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
		store                *storage.Store // Initial state of the data store prior to HTTP request
		postBody             testRequest    // HTTP request body (in struct form)
		wantStatus           int            // Expected HTTP response code
		wantResponseContains []string       // For validating errors
		wantResponse         *testResponse  // For validating successes
		wantStore            *storage.Store // Expected state of the data store after HTTP request is complete
	}{
		{
			name: "Simple Assignment goes to first available agent",
			store: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{}},
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
			wantStore: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{}},
			}, nil),
		},
		{
			name: "Simple Assignment w/ existing task (1) goes to next available agent",
			store: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{}},
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
			wantStore: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{
					&storage.Task{ID: 2, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
			}, nil),
		},
		{
			name: "Simple Assignment w/ existing tasks (2) goes to last available agent",
			store: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{
					&storage.Task{ID: 2, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
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
			wantStore: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{
					&storage.Task{ID: 3, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill3}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{
					&storage.Task{ID: 2, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
			}, nil),
		},
		{
			name:                 "Assignment fails: no agent w/ skills available",
			store:                storage.NewStore([]*storage.Agent{}, nil),
			postBody:             testRequest{Priority: "high", ReqSkills: []string{"skill1"}},
			wantStatus:           http.StatusConflict,
			wantResponseContains: []string{`{"error":"Could not assign task: No existing agents possess the required skills for this task"}`},
			wantStore:            storage.NewStore([]*storage.Agent{}, nil),
		},
		{
			name: "Assignment fails: no agent available for priority",
			store: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{
					&storage.Task{ID: 2, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
			}, nil),
			postBody:             testRequest{Priority: "high", ReqSkills: []string{"skill1"}},
			wantStatus:           http.StatusConflict,
			wantResponseContains: []string{`{"error":"Could not assign task: No agents are currently available for this task priority"}`},
			wantStore: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{
					&storage.Task{ID: 2, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
			}, nil),
		},
		{
			name: "Assignment of higher priority proceeds to agent w/ most recently assigned task",
			store: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityLow, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP, AssignmentTime: time.Now().Add(-2 * time.Hour)},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{
					&storage.Task{ID: 2, Priority: storage.PriorityLow, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP, AssignmentTime: time.Now().Add(-1 * time.Hour)},
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
			wantStore: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityLow, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{
					&storage.Task{ID: 3, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
					&storage.Task{ID: 2, Priority: storage.PriorityLow, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
			}, nil),
		},
		{
			name: "Assignment of higher priority proceeds to agent w/ most recently assigned task, excluding other busy agents",
			store: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityLow, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP, AssignmentTime: time.Now().Add(-2 * time.Hour)},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{
					&storage.Task{ID: 3, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
					&storage.Task{ID: 2, Priority: storage.PriorityLow, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
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
			wantStore: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 4, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
					&storage.Task{ID: 1, Priority: storage.PriorityLow, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{
					&storage.Task{ID: 3, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
					&storage.Task{ID: 2, Priority: storage.PriorityLow, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
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
				var wantTask storage.Task
				err = json.Unmarshal(wantRespJSON, &wantTask) // Unmarshal expected test result JSON --> Task{}
				if err != nil {
					t.Fatal(err)
				}
				var gotTask storage.Task
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
		store      *storage.Store // Initial state of the data store prior to HTTP request
		postBody   testRequest    // HTTP request body (in struct form)
		wantStatus int            // Expected HTTP response code
		wantStore  *storage.Store // Expected state of the data store after HTTP request is complete
	}{
		{
			name: "Simple task completion",
			store: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{}},
			}, nil),
			postBody:   testRequest{ID: 1},
			wantStatus: http.StatusOK,
			wantStore: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{}},
			}, []*storage.Task{
				&storage.Task{ID: 1, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskComplete, AssignedAgent: &storage.Agent{ID: 1, Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}}},
			}),
		},
		{
			name: "Task completion with >1 tasks in queue",
			store: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 1, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
					&storage.Task{ID: 2, Priority: storage.PriorityLow, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{}},
			}, nil),
			postBody:   testRequest{ID: 1},
			wantStatus: http.StatusOK,
			wantStore: storage.NewStore([]*storage.Agent{
				&storage.Agent{Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}, Tasks: []*storage.Task{
					&storage.Task{ID: 2, Priority: storage.PriorityLow, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskInWIP},
				}},
				&storage.Agent{Name: "Betty", Skills: storage.Skills{storage.Skill2, storage.Skill3}, Tasks: []*storage.Task{}},
				&storage.Agent{Name: "Charlie", Skills: storage.Skills{storage.Skill1}, Tasks: []*storage.Task{}},
			}, []*storage.Task{
				&storage.Task{ID: 1, Priority: storage.PriorityHigh, ReqSkills: storage.Skills{storage.Skill1}, State: storage.TaskComplete, AssignedAgent: &storage.Agent{ID: 1, Name: "Adam", Skills: storage.Skills{storage.Skill1, storage.Skill2}}},
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
