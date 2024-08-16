package handlers

import (
	"net/http"

	"github.com/onmetal-dev/metal/cmd/app/templates"
)

type AboutHandLer struct{}

func NewAboutHandler() *AboutHandLer {
	return &AboutHandLer{}
}

func (h *AboutHandLer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := templates.About()
	err := templates.Layout(c, "metal | about").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
