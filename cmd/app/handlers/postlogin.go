package handlers

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/onmetal-dev/metal/cmd/app/hash"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/store"
)

func init() {
	// in order to save custom structs in the session, we need to register them
	gob.Register(store.User{})
}

type PostLoginHandler struct {
	userStore    store.UserStore
	teamStore    store.TeamStore
	passwordhash hash.PasswordHash
	sessionStore sessions.Store
	sessionName  string
}

func NewPostLoginHandler(
	userStore store.UserStore,
	teamStore store.TeamStore,
	passwordHash hash.PasswordHash,
	sessionStore sessions.Store,
	sessionName string,
) *PostLoginHandler {
	return &PostLoginHandler{
		userStore:    userStore,
		teamStore:    teamStore,
		passwordhash: passwordHash,
		sessionStore: sessionStore,
		sessionName:  sessionName,
	}
}

func (h *PostLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	next := r.FormValue("next")
	user, err := h.userStore.GetUser(email)
	if err != nil || user == nil {
		logger.FromContext(r.Context()).Error("failed to get user with email", "email", email, "error", err)
		w.WriteHeader(http.StatusUnauthorized)
		c := templates.LoginError()
		c.Render(r.Context(), w)
		return
	}

	passwordIsValid, err := h.passwordhash.ComparePasswordAndHash(password, user.Password)

	if err != nil || !passwordIsValid {
		logger.FromContext(r.Context()).Error("failed to compare password and hash", "email", email, "error", err)
		w.WriteHeader(http.StatusUnauthorized)
		c := templates.LoginError()
		c.Render(r.Context(), w)
		return
	}

	session, _ := h.sessionStore.Get(r, h.sessionName)
	logger.FromContext(r.Context()).Info("user logged in", "user", user)
	session.Values["user"] = user
	if err := h.sessionStore.Save(r, w, session); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.FromContext(r.Context()).Info("user", "user", user)

	// first log in / onboarding logic
	if len(user.TeamMemberships) == 1 {
		logger.FromContext(r.Context()).Info("user has one team membership", "user", user)
		// check for incomplete onboarding
		team, err := h.teamStore.GetTeam(user.TeamMemberships[0].TeamID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(team.PaymentMethods) == 0 {
			w.Header().Set("HX-Redirect", fmt.Sprintf("/onboarding/%s/payment", team.ID))
			w.WriteHeader(http.StatusOK)
			return
		}
	} else if len(user.TeamMemberships) == 0 {
		// check if they've been invited to a team
		invites, err := h.teamStore.GetInvitesForEmail(user.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(invites) > 0 {
			// auto accept the invites and redirect to /dashboard/{first teamId}
			for _, invite := range invites {
				if err := h.teamStore.AddUserToTeam(invite.TeamID, user.ID); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			firstTeamId := invites[0].TeamID
			w.Header().Set("HX-Redirect", "/dashboard/"+firstTeamId)
			w.WriteHeader(http.StatusOK)
			return
		}

		// at this point the user has no team memberships and has not been invited to a team,
		// so put them through the onboarding flow that creates a team
		w.Header().Set("HX-Redirect", "/onboarding")
		w.WriteHeader(http.StatusOK)
		return
	}

	// can assume now that user is part of a team and that team has payment set up
	if next != "" {
		w.Header().Set("HX-Redirect", next)
	} else {
		w.Header().Set("HX-Redirect", "/dashboard/"+user.TeamMemberships[0].TeamID)
	}
	w.WriteHeader(http.StatusOK)
}
