package handlers

import (
	"net/http"

	"github.com/onmetal-dev/metal/cmd/app/templates"
)

type HomeHandler struct {
}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := templates.Index()
	if err := templates.Layout(c, "metal").Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
