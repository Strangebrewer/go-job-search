package recruiter

import "github.com/go-chi/chi/v5"

func Routes(store *Store) chi.Router {
	r := chi.NewRouter()
	h := NewHandler(store)

	r.Get("/", h.List)
	r.Get("/{id}", h.GetOne)
	r.Post("/", h.Create)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}
