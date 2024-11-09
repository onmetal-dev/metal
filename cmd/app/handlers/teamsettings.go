package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/store"
)

type GetTeamSettingsHandler struct {
	userStore     store.UserStore
	teamStore     store.TeamStore
	apiTokenStore store.ApiTokenStore
}

func NewGetTeamSettingsHandler(userStore store.UserStore, teamStore store.TeamStore, apiTokenStore store.ApiTokenStore) *GetTeamSettingsHandler {
	return &GetTeamSettingsHandler{
		userStore:     userStore,
		teamStore:     teamStore,
		apiTokenStore: apiTokenStore,
	}
}

func (h *GetTeamSettingsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, userTeams := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}

	apiTokens, err := h.apiTokenStore.List(teamId)
	if err != nil {
		http.Error(w, "Error fetching API keys", http.StatusInternalServerError)
		return
	}

	dashboardState := templates.DashboardState{
		User:              *user,
		Teams:             userTeams,
		ActiveTeam:        *team,
		AdditionalScripts: []templates.ScriptTag{},
	}

	if err := templates.DashboardLayout(dashboardState, templates.TeamSettings(teamId, *team, apiTokens)).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}
