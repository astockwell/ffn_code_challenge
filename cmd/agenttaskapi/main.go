package main

import (
	"net/http"

	"github.com/astockwell/ffn/pkg/storage"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/unrolled/render"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func main() {
	// Setup Logging
	log.SetFormatter(&prefixed.TextFormatter{FullTimestamp: true, TimestampFormat: "2006/01/02 15:04:05", ForceFormatting: true})
	log.SetLevel(log.TraceLevel)

	// Setup data store
	store := &storage.Store{}

	// Seed data store
	log.Tracef("Building Seed Agents...")
	agents := storage.BuildSeedAgents()
	log.Tracef("Persisting Seed Agents...")
	err := store.AddAgents(agents)
	if err != nil {
		log.Fatal("Error provisioning seed Agents:", err)
	}

	// Prepare web server components
	renderer := render.New()
	router := httprouter.New()
	dso := &DataSourceOrchestration{
		Renderer: renderer,
		Store:    store,
	}

	// Web server routes
	router.GET("/", mwLogger(route_Index(dso)))
	router.POST("/tasks/new", mwLogger(route_Tasks_New_POST(dso)))
	router.POST("/tasks/complete", mwLogger(route_Tasks_Update_Complete_POST(dso)))

	// Serve HTTP
	log.Infof("HTTP Web server (no TLS) listening on %s", ":8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
