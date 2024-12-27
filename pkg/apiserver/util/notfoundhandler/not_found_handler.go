package notfoundhandler

import "net/http"

func New() *Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	http.NotFound(rw, req)
}
