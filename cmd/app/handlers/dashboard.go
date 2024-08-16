package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/store"
)

type DashboardHandler struct {
	userStore   store.UserStore
	teamStore   store.TeamStore
	serverStore store.ServerStore
}

func NewDashboardHandler(userStore store.UserStore, teamStore store.TeamStore, serverStore store.ServerStore) *DashboardHandler {
	return &DashboardHandler{
		userStore:   userStore,
		teamStore:   teamStore,
		serverStore: serverStore,
	}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(r.Context())
	team, userTeams := validateAndFetchTeams(h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	servers, err := h.serverStore.GetServersForTeam(teamId)
	if err != nil {
		http.Error(w, "error fetching servers", http.StatusInternalServerError)
		return
	}

	if err := templates.DashboardLayout(templates.DashboardState{
		User:          *user,
		UserTeams:     userTeams,
		ActiveTeam:    *team,
		ActiveTabName: templates.TabNameHome,
	}, templates.DashboardHome(teamId, servers)).Render(r.Context(), w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}
