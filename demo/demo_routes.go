package demo

import (
	"github.com/go-chi/chi/v5"

	"github.com/Strangebrewer/go-job-search/job"
	"github.com/Strangebrewer/go-job-search/middleware"
	"github.com/Strangebrewer/go-job-search/recruiter"
	"github.com/Strangebrewer/go-job-search/tracer"
)

func Routes(recruiterStore *recruiter.Store, jobStore *job.Store, tc *tracer.Client, audience string) chi.Router {
	r := chi.NewRouter()
	h := NewHandler(recruiterStore, jobStore, tc)
	r.With(middleware.RequirePubSubOIDC(audience)).Post("/demo-registered", h.HandleDemoRegistered)
	return r
}
