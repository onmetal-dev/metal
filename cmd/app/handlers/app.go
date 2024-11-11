package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/cmd/app/urls"
	"github.com/onmetal-dev/metal/lib/background"
	"github.com/onmetal-dev/metal/lib/background/deployment"
	"github.com/onmetal-dev/metal/lib/form"
	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

type AppDetailsHandler struct {
	userStore          store.UserStore
	teamStore          store.TeamStore
	serverStore        store.ServerStore
	cellStore          store.CellStore
	deploymentStore    store.DeploymentStore
	appStore           store.AppStore
	producerDeployment *background.QueueProducer[deployment.Message]
}

func NewAppDetailsHandler(userStore store.UserStore, teamStore store.TeamStore, serverStore store.ServerStore, cellStore store.CellStore, deploymentStore store.DeploymentStore, appStore store.AppStore, producerDeployment *background.QueueProducer[deployment.Message]) *AppDetailsHandler {
	return &AppDetailsHandler{
		userStore:          userStore,
		teamStore:          teamStore,
		serverStore:        serverStore,
		cellStore:          cellStore,
		deploymentStore:    deploymentStore,
		appStore:           appStore,
		producerDeployment: producerDeployment,
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
	latestDeployment, err := h.deploymentStore.GetLatestForAppEnv(ctx, appId, env.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var f templates.UpdateAppEnvVarsFormData
	if len(latestDeployment.AppEnvVars.EnvVars.Data()) > 0 {
		envVarStr := lo.Associate(latestDeployment.AppEnvVars.EnvVars.Data(), func(ev store.EnvVar) (string, string) {
			return ev.Name, ev.Value
		})
		f.EnvVars, err = godotenv.Marshal(envVarStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := templates.DashboardLayout(templates.DashboardState{
		User:       *user,
		Teams:      teams,
		ActiveTeam: *team,
		Envs:       team.Envs,
		ActiveEnv:  env,
	}, templates.AppDetailsLayout(*team, *env, latestDeployment.App, templates.AppMenuItemVariables,
		templates.AppDetailsVariables(teamId, env.Name, appId, f, form.FieldErrors{}, nil))).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}

func (h *AppDetailsHandler) ServeHTTPVariablesUpdate(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	envName := chi.URLParam(r, "envName")
	appId := chi.URLParam(r, "appId")
	user := middleware.GetUser(ctx)
	team, _ := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
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

	var f templates.UpdateAppEnvVarsFormData
	inputErrs, err := form.Decode(&f, r)
	if inputErrs.NotNil() || err != nil {
		// send back the form html w/ errors
		if err := templates.UpdateAppEnvVarsForm(teamId, envName, appId, f, inputErrs, err).Render(ctx, w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	var envVars []store.EnvVar
	if f.EnvVars != "" {
		parsedEnvVars, err := godotenv.Parse(strings.NewReader(f.EnvVars))
		if err != nil {
			http.Error(w, fmt.Sprintf("error parsing environment variables: %v", err), http.StatusBadRequest)
			return
		}
		for k, v := range parsedEnvVars {
			envVars = append(envVars, store.EnvVar{Name: k, Value: v})
		}
	}
	appEnvVars, err := h.deploymentStore.CreateAppEnvVars(store.CreateAppEnvVarOptions{
		TeamId:  teamId,
		EnvId:   env.Id,
		AppId:   appId,
		EnvVars: envVars,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating app env vars: %v", err), http.StatusInternalServerError)
		return
	}

	// create a new deployment using the latest deployment as a template
	latestDeployment, err := h.deploymentStore.GetLatestForAppEnv(ctx, appId, env.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Info("creating deployment")
	d, err := h.deploymentStore.Create(store.CreateDeploymentOptions{
		TeamId:        teamId,
		Type:          store.DeploymentTypeDeploy,
		EnvId:         env.Id,
		AppId:         appId,
		AppSettingsId: latestDeployment.AppSettingsId,
		AppEnvVarsId:  appEnvVars.Id,
		CellIds:       lo.Map(latestDeployment.Cells, func(c store.Cell, _ int) string { return c.Id }),
		Replicas:      latestDeployment.Replicas,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send a message to the deployment queue
	err = h.producerDeployment.Send(ctx, deployment.Message{
		DeploymentId: d.Id,
		AppId:        d.AppId,
		EnvId:        d.EnvId,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// redirect to /deployments page
	middleware.AddFlash(ctx, fmt.Sprintf("environment variables updated successfully. deployment %d created", d.Id))
	w.Header().Set("HX-Redirect", urls.EnvAppDeployments{TeamId: teamId, AppId: appId, EnvName: envName}.Render())
	w.WriteHeader(http.StatusOK)
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
	latestDeployment, err := h.deploymentStore.GetLatestForAppEnv(ctx, appId, env.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.DashboardLayout(templates.DashboardState{
		User:       *user,
		Teams:      teams,
		ActiveTeam: *team,
		Envs:       team.Envs,
		ActiveEnv:  env,
	}, templates.AppDetailsLayout(*team, *env, latestDeployment.App, templates.AppMenuItemSettings, templates.AppDetailsSettings(latestDeployment.AppSettings))).Render(ctx, w); err != nil {
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
			sortedDeployments = lo.Filter(sortedDeployments, func(d store.Deployment, _ int) bool {
				return d.Id != activeDeployment.Id
			})
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
