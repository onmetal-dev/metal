package handlers

import (
	"net/http"

	"github.com/onmetal-dev/metal/cmd/app/templates"
)

type NotFoundHandler struct{}

func NewNotFoundHandler() *NotFoundHandler {
	return &NotFoundHandler{}
}

func (h *NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := templates.NotFound()
	if err := templates.LayoutBare(c, "oops, nothing to see here").Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
