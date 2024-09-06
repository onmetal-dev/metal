package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/store"
	"golang.org/x/sync/errgroup"
)

type DashboardHandler struct {
	userStore           store.UserStore
	teamStore           store.TeamStore
	serverStore         store.ServerStore
	cellStore           store.CellStore
	deploymentStore     store.DeploymentStore
	appStore            store.AppStore
	cellProviderForType func(cellType store.CellType) cellprovider.CellProvider
}

func NewDashboardHandler(userStore store.UserStore, teamStore store.TeamStore, serverStore store.ServerStore, cellStore store.CellStore, deploymentStore store.DeploymentStore, appStore store.AppStore, cellProviderForType func(cellType store.CellType) cellprovider.CellProvider) *DashboardHandler {
	return &DashboardHandler{
		userStore:           userStore,
		teamStore:           teamStore,
		serverStore:         serverStore,
		cellStore:           cellStore,
		deploymentStore:     deploymentStore,
		appStore:            appStore,
		cellProviderForType: cellProviderForType,
	}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, userTeams := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	var (
		servers     []store.Server
		cells       []store.Cell
		serverStats []cellprovider.ServerStats
		deployments []store.Deployment
		apps        []store.App
	)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		servers, err = h.serverStore.GetServersForTeam(ctx, teamId)
		return err
	})

	g.Go(func() error {
		var err error
		cells, err = h.cellStore.GetForTeam(ctx, teamId)
		return err
	})

	g.Go(func() error {
		var err error
		cells, err = h.cellStore.GetForTeam(ctx, teamId)
		if err != nil {
			return err
		}

		var stats []cellprovider.ServerStats
		for _, cell := range cells {
			cellStats, err := h.cellProviderForType(cell.Type).ServerStats(ctx, cell.Id)
			if err != nil {
				return fmt.Errorf("error fetching server stats: %v", err)
			}
			stats = append(stats, cellStats...)
		}
		serverStats = stats
		return nil
	})

	g.Go(func() error {
		var err error
		deployments, err = h.deploymentStore.GetForTeam(ctx, teamId)
		return err
	})

	g.Go(func() error {
		var err error
		apps, err = h.appStore.GetForTeam(ctx, teamId)
		return err
	})

	if err := g.Wait(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.DashboardLayout(templates.DashboardState{
		User:          *user,
		UserTeams:     userTeams,
		ActiveTeam:    *team,
		ActiveTabName: templates.TabNameHome,
		AdditionalScripts: []templates.ScriptTag{
			templates.ScriptTag{
				Src: "/static/script/sse.js",
			},
		},
	}, templates.DashboardHome(teamId, servers, cells, serverStats, deployments, apps)).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}
