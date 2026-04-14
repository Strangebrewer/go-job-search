package job

import (
	"github.com/go-chi/chi/v5"

	"github.com/Strangebrewer/go-job-search/tracer"
)

func Routes(store *Store, tc *tracer.Client) chi.Router {
	r := chi.NewRouter()
	h := NewHandler(store, tc)

	r.Get("/", h.List)
	r.Get("/{id}", h.GetOne)
	r.Post("/", h.Create)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}
