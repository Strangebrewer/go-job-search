package job

import (
	"github.com/go-chi/chi/v5"

	"github.com/Strangebrewer/go-job-search/pubsub"
)

func Routes(store *Store, publisher *pubsub.Publisher, interviewScheduledTopicID string) chi.Router {
	r := chi.NewRouter()
	h := NewHandler(store, publisher, interviewScheduledTopicID)

	r.Get("/", h.List)
	r.Get("/{id}", h.GetOne)
	r.Post("/", h.Create)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}
