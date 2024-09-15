package handlers

import (
	"net/http"

	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/store"
)

type GetSignUpHandler struct {
	inviteStore store.InviteStore
	teamStore   store.TeamStore
}

func NewGetSignUpHandler(inviteStore store.InviteStore, teamStore store.TeamStore) *GetSignUpHandler {
	return &GetSignUpHandler{
		inviteStore: inviteStore,
		teamStore:   teamStore,
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
		if !allowed {
			invites, err := h.teamStore.GetInvitesForEmail(email)
			if err != nil {
				http.Error(w, "Error fetching invites", http.StatusInternalServerError)
				return
			}
			allowed = len(invites) > 0
		}
	}

	c := templates.SignUpPage(allowed)
	if err := templates.Layout(c, "metal").Render(r.Context(), w); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
