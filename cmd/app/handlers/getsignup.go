package handlers

import (
	"net/http"

	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/store"
)

type GetSignUpHandler struct {
	inviteStore store.InviteStore
}

func NewGetSignUpHandler(inviteStore store.InviteStore) *GetSignUpHandler {
	return &GetSignUpHandler{
		inviteStore: inviteStore,
	}
}

func (h *GetSignUpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var allowed bool
	email := r.URL.Query().Get("email")
	if email != "" {
		invitedUser, err := h.inviteStore.Get(email)
		if err != nil {
			http.Error(w, "Error fetching invited user", http.StatusInternalServerError)
			return
		}
		allowed = invitedUser != nil
	}

	c := templates.SignUpPage(allowed)
	if err := templates.Layout(c, "metal").Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
