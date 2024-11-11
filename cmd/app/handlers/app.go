package handlers

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/cmd/app/urls"
	"github.com/onmetal-dev/metal/lib/store"
	"golang.org/x/sync/errgroup"
)

type AppDetailsHandler struct {
	userStore       store.UserStore
	teamStore       store.TeamStore
	serverStore     store.ServerStore
	cellStore       store.CellStore
	deploymentStore store.DeploymentStore
	appStore        store.AppStore
}

func NewAppDetailsHandler(userStore store.UserStore, teamStore store.TeamStore, serverStore store.ServerStore, cellStore store.CellStore, deploymentStore store.DeploymentStore, appStore store.AppStore) *AppDetailsHandler {
	return &AppDetailsHandler{
		userStore:       userStore,
		teamStore:       teamStore,
		serverStore:     serverStore,
		cellStore:       cellStore,
		deploymentStore: deploymentStore,
		appStore:        appStore,
	}
}

func (h *AppDetailsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	teamId := chi.URLParam(r, "teamId")
	envName := chi.URLParam(r, "envName")
	appId := chi.URLParam(r, "appId")
	// redirect to root to /deployments
	http.Redirect(w, r, urls.EnvAppDeployments{TeamId: teamId, AppId: appId, EnvName: envName}.Render(), http.StatusTemporaryRedirect)
}

func (h *AppDetailsHandler) ServeHTTPVariables(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	envName := chi.URLParam(r, "envName")
	appId := chi.URLParam(r, "appId")
	user := middleware.GetUser(ctx)
	team, teams := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}

	var env *store.Env
	for _, e := range team.Envs {
		if e.Name == envName {
			env = &e
		}
	}
	if env == nil {
		http.Error(w, "env not found", http.StatusNotFound)
		return
	}

	var (
		app *store.App
	)

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		a, err := h.appStore.Get(ctx, appId)
		if err != nil {
			return err
		}
		app = &a
		return nil
	})

	if err := g.Wait(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.DashboardLayout(templates.DashboardState{
		User:       *user,
		Teams:      teams,
		ActiveTeam: *team,
		Envs:       team.Envs,
		ActiveEnv:  env,
	}, templates.AppDetailsLayout(*team, *env, *app, templates.AppMenuItemVariables, templates.AppDetailsVariables(*app))).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}

func (h *AppDetailsHandler) ServeHTTPSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	envName := chi.URLParam(r, "envName")
	appId := chi.URLParam(r, "appId")
	user := middleware.GetUser(ctx)
	team, teams := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}

	var env *store.Env
	for _, e := range team.Envs {
		if e.Name == envName {
			env = &e
		}
	}
	if env == nil {
		http.Error(w, "env not found", http.StatusNotFound)
		return
	}

	var (
		app *store.App
	)

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		a, err := h.appStore.Get(ctx, appId)
		if err != nil {
			return err
		}
		app = &a
		return nil
	})

	if err := g.Wait(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.DashboardLayout(templates.DashboardState{
		User:       *user,
		Teams:      teams,
		ActiveTeam: *team,
		Envs:       team.Envs,
		ActiveEnv:  env,
	}, templates.AppDetailsLayout(*team, *env, *app, templates.AppMenuItemSettings, templates.AppDetailsSettings(*app))).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}

func byDateDescending(deployments []store.Deployment) []store.Deployment {
	sort.Slice(deployments, func(i, j int) bool {
		return deployments[i].CreatedAt.After(deployments[j].CreatedAt)
	})
	return deployments
}

func (h *AppDetailsHandler) ServeHTTPDeployments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	envName := chi.URLParam(r, "envName")
	appId := chi.URLParam(r, "appId")
	user := middleware.GetUser(ctx)
	team, teams := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}

	var env *store.Env
	for _, e := range team.Envs {
		if e.Name == envName {
			env = &e
		}
	}
	if env == nil {
		http.Error(w, "env not found", http.StatusNotFound)
		return
	}

	var (
		deployments []store.Deployment
		app         *store.App
	)

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		deployments, err = h.deploymentStore.GetForAppEnv(ctx, appId, env.Id)
		return err
	})

	g.Go(func() error {
		a, err := h.appStore.Get(ctx, appId)
		if err != nil {
			return err
		}
		app = &a
		return nil
	})

	if err := g.Wait(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sortedDeployments := byDateDescending(deployments)

	var activeDeployment *store.Deployment
	for _, deployment := range sortedDeployments {
		if deployment.Status == store.DeploymentStatusRunning {
			activeDeployment = &deployment
			break
		}
	}

	if err := templates.DashboardLayout(templates.DashboardState{
		User:       *user,
		Teams:      teams,
		ActiveTeam: *team,
		Envs:       team.Envs,
		ActiveEnv:  env,
	}, templates.AppDetailsLayout(*team, *env, *app, templates.AppMenuItemDeployments, templates.AppDetailsDeployments(activeDeployment, sortedDeployments))).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}
