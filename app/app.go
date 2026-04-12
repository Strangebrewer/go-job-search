package app

import (
	"github.com/Strangebrewer/go-job-search/job"
	"github.com/Strangebrewer/go-job-search/recruiter"
)

type Application struct {
	JobStore       *job.Store
	RecruiterStore *recruiter.Store
}
