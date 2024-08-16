package handlers

import (
	"net/http"

	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/store"
)

type PostWaitlistHandler struct {
	waitlistStore store.WaitlistStore
}

type PostWaitlistHandlerParams struct {
	WaitlistStore store.WaitlistStore
}

func NewPostWaitlistHandler(params PostWaitlistHandlerParams) *PostWaitlistHandler {
	return &PostWaitlistHandler{
		waitlistStore: params.WaitlistStore,
	}
}

func (h *PostWaitlistHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	if err := h.waitlistStore.Add(email); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c := templates.WaitlistError(err.Error())
		c.Render(r.Context(), w)
		return
	}

	if err := templates.WaitlistSuccess().Render(r.Context(), w); err != nil {
		http.Error(w, "error rendering template", http.StatusInternalServerError)
		return
	}

}
