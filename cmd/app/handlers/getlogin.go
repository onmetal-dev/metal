package handlers

import (
	"net/http"

	"github.com/onmetal-dev/metal/cmd/app/templates"
)

type GetLoginHandler struct{}

func NewGetLoginHandler() *GetLoginHandler {
	return &GetLoginHandler{}
}

func (h *GetLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	next := r.URL.Query().Get("next")
	c := templates.Login(next)
	err := templates.Layout(c, "login to metal").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "error rendering template", http.StatusInternalServerError)
		return
	}
}
