package rube

import (
	"github.com/Strangebrewer/go-job-search/pubsub"
	"github.com/Strangebrewer/go-job-search/tracer"
	"github.com/go-chi/chi/v5"
)

func Routes(tc *tracer.Client, publisher *pubsub.Publisher, rubeOwidTopicId string) chi.Router {
	r := chi.NewRouter()
	h := NewHandler(tc, publisher, rubeOwidTopicId)
	r.Post("/", h.Chain)
	return r
}
