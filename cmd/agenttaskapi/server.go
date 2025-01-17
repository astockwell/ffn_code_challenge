package main

import (
	"github.com/astockwell/ffn/pkg/service"
	"github.com/unrolled/render"
)

// DataSourceOrchestration facilitates passing connection handles, etc,
// to handlers to prevent having constant propogating changes.
type DataSourceOrchestration struct {
	Renderer *render.Render
	Store    *service.Store
}
