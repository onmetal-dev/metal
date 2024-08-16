package handlers

import (
	"net/http"

	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/store"
)

type PostSignUpHandler struct {
	userStore   store.UserStore
	inviteStore store.InviteStore
	teamStore   store.TeamStore
}

func NewPostSignUpHandler(userStore store.UserStore, inviteStore store.InviteStore, teamStore store.TeamStore) *PostSignUpHandler {
	return &PostSignUpHandler{
		userStore:   userStore,
		inviteStore: inviteStore,
		teamStore:   teamStore,
	}
}

func (h *PostSignUpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	// verify user is either (1) off the waitlist or (2) an existing user invited them to their team
	allowed := false
	if invite, err := h.inviteStore.Get(email); err != nil {
		http.Error(w, "error checking invite", http.StatusInternalServerError)
		return
	} else if invite != nil {
		allowed = true
	} else if teamInvites, err := h.teamStore.GetInvitesForEmail(email); err != nil {
		http.Error(w, "error checking team invite", http.StatusInternalServerError)
		return
	} else if len(teamInvites) > 0 {
		allowed = true
	}
	if !allowed {
		http.Error(w, "user is not allowed to sign up", http.StatusForbidden)
		return
	}

	if err := h.userStore.CreateUser(email, password); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c := templates.SignUpError()
		c.Render(r.Context(), w)
		return
	}

	if err := templates.SignUpSuccess().Render(r.Context(), w); err != nil {
		http.Error(w, "error rendering template", http.StatusInternalServerError)
		return
	}

}
