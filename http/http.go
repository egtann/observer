package http

import "net/http"

type Server struct{}

type handler struct{}

func NewServer() *Server {
	h := &handler{}
	r := chi.NewRouter()
	r.Get("/", h.overview)
	r.Get("/roles", h.roles)
	r.Get("/hosts", h.hosts)
	r.Get("/request", h.request)
}

func (h *handler) overview(w http.ResponseWriter, r *http.Request) {
}
