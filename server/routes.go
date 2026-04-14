package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Strangebrewer/go-job-search/app"
	"github.com/Strangebrewer/go-job-search/health"
	"github.com/Strangebrewer/go-job-search/job"
	"github.com/Strangebrewer/go-job-search/recruiter"
)

func registerRoutes(r chi.Router, application *app.Application, authMiddleware func(http.Handler) http.Handler) {
	r.Get("/health", health.Handler)

	r.With(authMiddleware).Mount("/jobs", job.Routes(application.JobStore, application.Tracer))
	r.With(authMiddleware).Mount("/recruiters", recruiter.Routes(application.RecruiterStore))
}
