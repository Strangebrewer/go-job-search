package demo

import (
	"github.com/go-chi/chi/v5"

	"github.com/Strangebrewer/go-job-search/job"
	"github.com/Strangebrewer/go-job-search/middleware"
	"github.com/Strangebrewer/go-job-search/recruiter"
)

func Routes(recruiterStore *recruiter.Store, jobStore *job.Store, audience string) chi.Router {
	r := chi.NewRouter()
	h := NewHandler(recruiterStore, jobStore)
	r.With(middleware.RequirePubSubOIDC(audience)).Post("/demo-registered", h.HandleDemoRegistered)
	return r
}
