package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/background"
	"github.com/onmetal-dev/metal/lib/background/deployment"
	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/store"
)

type AppsNewHandler struct {
	userStore   store.UserStore
	teamStore   store.TeamStore
	serverStore store.ServerStore
	cellStore   store.CellStore
}

func NewGetAppsNewHandler(userStore store.UserStore, teamStore store.TeamStore, serverStore store.ServerStore, cellStore store.CellStore) *AppsNewHandler {
	return &AppsNewHandler{
		userStore:   userStore,
		teamStore:   teamStore,
		serverStore: serverStore,
		cellStore:   cellStore,
	}
}

func (h *AppsNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(r.Context())
	team, userTeams := validateAndFetchTeams(h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	cells, err := h.cellStore.GetForTeam(teamId)
	if err != nil {
		http.Error(w, "error fetching cells", http.StatusInternalServerError)
		return
	}

	dashboardState := templates.DashboardState{
		User:              *user,
		UserTeams:         userTeams,
		ActiveTeam:        *team,
		ActiveTabName:     templates.TabNameCreateApp,
		AdditionalScripts: []templates.ScriptTag{},
	}

	if len(cells) == 0 {
		if err := templates.DashboardLayout(dashboardState, templates.DashboardHomeNoServers(teamId)).Render(r.Context(), w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if err := templates.DashboardLayout(dashboardState, templates.CreateApp(teamId, cells, templates.CreateAppFormData{}, templates.CreateAppFormErrors{}, nil)).Render(r.Context(), w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}

type PostAppsNewHandler struct {
	userStore          store.UserStore
	teamStore          store.TeamStore
	serverStore        store.ServerStore
	cellStore          store.CellStore
	appStore           store.AppStore
	deploymentStore    store.DeploymentStore
	producerDeployment *background.QueueProducer[deployment.Message]
}

func NewPostAppsNewHandler(
	userStore store.UserStore,
	teamStore store.TeamStore,
	serverStore store.ServerStore,
	cellStore store.CellStore,
	appStore store.AppStore,
	deploymentStore store.DeploymentStore,
	producerDeployment *background.QueueProducer[deployment.Message],
) *PostAppsNewHandler {
	return &PostAppsNewHandler{
		userStore:          userStore,
		teamStore:          teamStore,
		serverStore:        serverStore,
		cellStore:          cellStore,
		appStore:           appStore,
		deploymentStore:    deploymentStore,
		producerDeployment: producerDeployment,
	}
}

func (h *PostAppsNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(r.Context())
	team, _ := validateAndFetchTeams(h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	cells, err := h.cellStore.GetForTeam(teamId)
	if err != nil {
		http.Error(w, "error fetching cells", http.StatusInternalServerError)
		return
	}
	f, inputErrs, err := templates.ParseCreateAppFormData(r)
	if inputErrs.NotNil() || err != nil {
		// send back the form html w/ errors
		if err := templates.CreateAppForm(teamId, cells, f, inputErrs, err).Render(r.Context(), w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// 1. Validate that the app name does not already exist
	existingApps, err := h.appStore.GetForTeam(teamId)
	if err != nil {
		http.Error(w, "error fetching existing apps", http.StatusInternalServerError)
		return
	}
	for _, app := range existingApps {
		if app.Name == *f.AppName {
			inputErrs.Set("AppName", fmt.Errorf("an app with this name already exists"))
			if err := templates.CreateAppForm(teamId, cells, f, inputErrs, nil).Render(r.Context(), w); err != nil {
				http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
			}
			return
		}
	}

	// 2. Validate that the cell ID exists and is part of the team
	var cellFound bool
	for _, cell := range cells {
		if cell.Id == *f.CellId {
			cellFound = true
			break
		}
	}
	if !cellFound {
		inputErrs.Set("CellId", fmt.Errorf("invalid cell ID"))
		if err := templates.CreateAppForm(teamId, cells, f, inputErrs, nil).Render(r.Context(), w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// 3. Create app, appenv, env, and deployment objects
	log.Info("creating app")
	app, err := h.appStore.Create(store.CreateAppOptions{
		Name:   *f.AppName,
		TeamId: teamId,
		UserId: user.Id,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating app: %v", err), http.StatusInternalServerError)
		return
	}

	// Check for existing "dev" environment or create it
	envs, err := h.deploymentStore.GetEnvsForTeam(teamId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching environments: %v", err), http.StatusInternalServerError)
		return
	}
	var devEnv store.Env
	for _, env := range envs {
		if env.Name == "dev" {
			devEnv = env
			break
		}
	}
	if devEnv.Id == "" {
		devEnv, err = h.deploymentStore.CreateEnv(store.CreateEnvOptions{
			TeamId: teamId,
			Name:   "dev",
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("error creating dev environment: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Create AppSettings
	appSettings, err := h.appStore.CreateAppSettings(store.CreateAppSettingsOptions{
		TeamId: teamId,
		AppId:  app.Id,
		Artifact: store.Artifact{
			Image: store.Image{Name: *f.ContainerImage},
		},
		Ports: store.Ports{{
			Name:  "http",
			Port:  *f.ContainerPort,
			Proto: "http",
		}},
		ExternalPorts: store.ExternalPorts{},
		Resources: store.Resources{
			Limits: store.ResourceLimits{
				CpuCores:  *f.CpuLimit,
				MemoryMiB: *f.MemoryLimit,
			},
			Requests: store.ResourceRequests{
				CpuCores:  *f.CpuLimit,
				MemoryMiB: *f.MemoryLimit,
			},
		},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating app settings: %v", err), http.StatusInternalServerError)
		return
	}

	// Create AppEnvVars
	var envVars []store.EnvVar
	if f.EnvVars != nil {
		parsedEnvVars, err := godotenv.Parse(strings.NewReader(*f.EnvVars))
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
		EnvId:   devEnv.Id,
		AppId:   app.Id,
		EnvVars: envVars,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating app env vars: %v", err), http.StatusInternalServerError)
		return
	}

	// Create Deployment
	log.Info("creating deployment")
	d, err := h.deploymentStore.Create(store.CreateDeploymentOptions{
		TeamId:        teamId,
		EnvId:         devEnv.Id,
		AppId:         app.Id,
		Type:          store.DeploymentTypeDeploy,
		AppSettingsId: appSettings.Id,
		AppEnvVarsId:  appEnvVars.Id,
		CellIds:       []string{*f.CellId},
		Replicas:      *f.Replicas,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating deployment: %v", err), http.StatusInternalServerError)
		return
	}

	// Send a message to the deployment queue
	err = h.producerDeployment.Send(r.Context(), deployment.Message{
		DeploymentId: d.Id,
		AppId:        d.AppId,
		EnvId:        d.EnvId,
	})
	if err != nil {
		log.Error("Failed to send deployment message to queue",
			slog.Any("error", err),
			slog.Int("deploymentId", int(d.Id)),
			slog.String("appId", d.AppId),
			slog.String("envId", d.EnvId),
		)
	}

	// Redirect to the dashboard on success
	middleware.AddFlash(r.Context(), fmt.Sprintf("app %s created successfully", *f.AppName))
	w.Header().Set("HX-Redirect", fmt.Sprintf("/dashboard/%s", teamId))
	w.WriteHeader(http.StatusOK)
}
