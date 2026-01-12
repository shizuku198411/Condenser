package http

import (
	_ "condenser/docs"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Condenser API
// @version 1.0
// @description High-level container runtime API for Raind stack
// @BasePath /
// @schemes http

func NewApiRouter() *chi.Mux {
	r := chi.NewRouter()
	handler := NewRequestHandler()

	// middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// == swagger ==
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	// == v1 ==
	// == containers ==
	r.Post("/v1/containers", handler.CreateContainer)                                // create container
	r.Post("/v1/containers/{containerId}/actions/start", handler.StartContainer)     // start container
	r.Post("/v1/containers/{containerId}/actions/stop", handler.StopContainer)       // stop container
	r.Post("/v1/containers/{containerId}/actions/exec", handler.ExecContainer)       // exec container
	r.Delete("/v1/containers/{containerId}/actions/delete", handler.DeleteContainer) // delete container

	return r
}

func NewHookRouter() *chi.Mux {
	r := chi.NewRouter()
	handler := NewRequestHandler()

	// middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// == hook ==
	r.Post("/v1/hooks/droplet", handler.ApplyHook)

	return r
}
