package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/store"
)

type DashboardHandler struct {
	userStore           store.UserStore
	teamStore           store.TeamStore
	serverStore         store.ServerStore
	cellStore           store.CellStore
	deploymentStore     store.DeploymentStore
	cellProviderForType func(cellType store.CellType) cellprovider.CellProvider
}

func NewDashboardHandler(userStore store.UserStore, teamStore store.TeamStore, serverStore store.ServerStore, cellStore store.CellStore, deploymentStore store.DeploymentStore, cellProviderForType func(cellType store.CellType) cellprovider.CellProvider) *DashboardHandler {
	return &DashboardHandler{
		userStore:           userStore,
		teamStore:           teamStore,
		serverStore:         serverStore,
		cellStore:           cellStore,
		deploymentStore:     deploymentStore,
		cellProviderForType: cellProviderForType,
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
	cells, err := h.cellStore.GetForTeam(teamId)
	if err != nil {
		http.Error(w, "error fetching cells", http.StatusInternalServerError)
		return
	}

	serverStats := []cellprovider.ServerStats{}
	for _, cell := range cells {
		stats, err := h.cellProviderForType(cell.Type).ServerStats(r.Context(), cell.Id)
		if err != nil {
			http.Error(w, fmt.Sprintf("error fetching server stats: %v", err), http.StatusInternalServerError)
			return
		}
		// for now assume these are ordered correctly (until we can figure out matching talos servers within the cell provider to our server object)
		serverStats = append(serverStats, stats...)
	}

	deployments, err := h.deploymentStore.GetForTeam(teamId)
	if err != nil {
		http.Error(w, "error fetching deployments", http.StatusInternalServerError)
		return
	}

	if err := templates.DashboardLayout(templates.DashboardState{
		User:          *user,
		UserTeams:     userTeams,
		ActiveTeam:    *team,
		ActiveTabName: templates.TabNameHome,
	}, templates.DashboardHome(teamId, servers, cells, serverStats, deployments)).Render(r.Context(), w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}
