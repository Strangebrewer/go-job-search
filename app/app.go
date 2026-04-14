package app

import (
	"github.com/Strangebrewer/go-job-search/job"
	"github.com/Strangebrewer/go-job-search/recruiter"
	"github.com/Strangebrewer/go-job-search/tracer"
)

type Application struct {
	JobStore       *job.Store
	RecruiterStore *recruiter.Store
	Tracer         *tracer.Client
}
