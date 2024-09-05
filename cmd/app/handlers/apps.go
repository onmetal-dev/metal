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
	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/form"
	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/samber/lo"
)

type AppsNewHandler struct {
	userStore   store.UserStore
	teamStore   store.TeamStore
	serverStore store.ServerStore
	cellStore   store.CellStore
}

func NewAppsNewHandler(userStore store.UserStore, teamStore store.TeamStore, serverStore store.ServerStore, cellStore store.CellStore) *AppsNewHandler {
	return &AppsNewHandler{
		userStore:   userStore,
		teamStore:   teamStore,
		serverStore: serverStore,
		cellStore:   cellStore,
	}
}

func (h *AppsNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, userTeams := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	cells, err := h.cellStore.GetForTeam(ctx, teamId)
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
		if err := templates.DashboardLayout(dashboardState, templates.DashboardHomeNoServers(teamId)).Render(ctx, w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if err := templates.DashboardLayout(dashboardState, templates.CreateApp(teamId, cells, templates.CreateAppFormData{}, form.FieldErrors{}, nil)).Render(ctx, w); err != nil {
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
	ctx := r.Context()
	log := logger.FromContext(ctx)
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, _ := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	cells, err := h.cellStore.GetForTeam(ctx, teamId)
	if err != nil {
		http.Error(w, "error fetching cells", http.StatusInternalServerError)
		return
	}
	var f templates.CreateAppFormData
	inputErrs, err := form.Decode(&f, r)
	if inputErrs.NotNil() || err != nil {
		// send back the form html w/ errors
		if err := templates.CreateAppForm(teamId, cells, f, inputErrs, err).Render(ctx, w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// 1. Validate that the app name does not already exist
	existingApps, err := h.appStore.GetForTeam(ctx, teamId)
	if err != nil {
		http.Error(w, "error fetching existing apps", http.StatusInternalServerError)
		return
	}
	for _, app := range existingApps {
		if app.Name == f.AppName {
			inputErrs.Set("AppName", fmt.Errorf("an app with this name already exists"))
			if err := templates.CreateAppForm(teamId, cells, f, inputErrs, nil).Render(ctx, w); err != nil {
				http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
			}
			return
		}
	}

	// 2. Validate that the cell ID exists and is part of the team
	var cellFound bool
	for _, cell := range cells {
		if cell.Id == f.CellId {
			cellFound = true
			break
		}
	}
	if !cellFound {
		inputErrs.Set("CellId", fmt.Errorf("invalid cell ID"))
		if err := templates.CreateAppForm(teamId, cells, f, inputErrs, nil).Render(ctx, w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// 3. Create app, appenv, env, and deployment objects
	log.Info("creating app")
	app, err := h.appStore.Create(store.CreateAppOptions{
		Name:   f.AppName,
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
			Image: store.Image{Name: f.ContainerImage},
		},
		Ports: store.Ports{{
			Name:  "http",
			Port:  f.ContainerPort,
			Proto: "http",
		}},
		ExternalPorts: store.ExternalPorts{},
		Resources: store.Resources{
			Limits: store.ResourceLimits{
				CpuCores:  f.CpuLimit,
				MemoryMiB: f.MemoryLimit,
			},
			Requests: store.ResourceRequests{
				CpuCores:  f.CpuLimit,
				MemoryMiB: f.MemoryLimit,
			},
		},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating app settings: %v", err), http.StatusInternalServerError)
		return
	}

	// Create AppEnvVars
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
		CellIds:       []string{f.CellId},
		Replicas:      f.Replicas,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating deployment: %v", err), http.StatusInternalServerError)
		return
	}

	// Send a message to the deployment queue
	err = h.producerDeployment.Send(ctx, deployment.Message{
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
	middleware.AddFlash(ctx, fmt.Sprintf("app %s created successfully", f.AppName))
	w.Header().Set("HX-Redirect", fmt.Sprintf("/dashboard/%s", teamId))
	w.WriteHeader(http.StatusOK)
}

type DeleteAppHandler struct {
	userStore           store.UserStore
	teamStore           store.TeamStore
	serverStore         store.ServerStore
	cellStore           store.CellStore
	appStore            store.AppStore
	deploymentStore     store.DeploymentStore
	cellProviderForType func(cellType store.CellType) cellprovider.CellProvider
}

func NewDeleteAppHandler(
	userStore store.UserStore,
	teamStore store.TeamStore,
	serverStore store.ServerStore,
	cellStore store.CellStore,
	appStore store.AppStore,
	deploymentStore store.DeploymentStore,
	cellProviderForType func(cellType store.CellType) cellprovider.CellProvider,
) *DeleteAppHandler {
	return &DeleteAppHandler{
		userStore:           userStore,
		teamStore:           teamStore,
		serverStore:         serverStore,
		cellStore:           cellStore,
		appStore:            appStore,
		deploymentStore:     deploymentStore,
		cellProviderForType: cellProviderForType,
	}
}

func (h *DeleteAppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, _ := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	cells, err := h.cellStore.GetForTeam(ctx, teamId)
	if err != nil {
		http.Error(w, "error fetching cells", http.StatusInternalServerError)
		return
	}

	appId := chi.URLParam(r, "appId")
	if appId == "" {
		http.Error(w, "appId is required", http.StatusBadRequest)
		return
	}

	app, err := h.appStore.Get(ctx, appId)
	if err != nil {
		http.Error(w, "error fetching app", http.StatusInternalServerError)
		return
	}

	if app.TeamId != teamId {
		http.Error(w, "app does not belong to team", http.StatusNotFound)
		return
	}

	// get all deployments for the app
	deployments, err := h.deploymentStore.GetForApp(ctx, app.Id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching deployments: %v", err), http.StatusInternalServerError)
		return
	}

	for _, cell := range cells {
		deploymentsForCell := lo.Filter(deployments, func(deployment store.Deployment, _ int) bool {
			inThisCell := false
			for _, c := range deployment.Cells {
				if c.Id == cell.Id {
					inThisCell = true
					break
				}
			}
			return inThisCell
		})
		if err := h.cellProviderForType(cell.Type).DestroyDeployments(ctx, cell.Id, deploymentsForCell); err != nil {
			http.Error(w, fmt.Sprintf("error destroying deployments: %v", err), http.StatusInternalServerError)
			return
		}
		// update deployment status for all deployments, then delete them so they don't appear in the dashboard
		for _, d := range deploymentsForCell {
			if err := h.deploymentStore.UpdateDeploymentStatus(app.Id, d.EnvId, d.Id, store.DeploymentStatusStopped, "app deleted"); err != nil {
				http.Error(w, fmt.Sprintf("error updating deployment status: %v", err), http.StatusInternalServerError)
				return
			}
			if err := h.deploymentStore.DeleteDeployment(app.Id, d.EnvId, d.Id); err != nil {
				http.Error(w, fmt.Sprintf("error deleting deployment: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}

	if err := h.appStore.Delete(ctx, app.Id); err != nil {
		http.Error(w, fmt.Sprintf("error deleting app: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to the dashboard on success
	middleware.AddFlash(ctx, fmt.Sprintf("app %s deleted successfully", app.Name))
	w.Header().Set("HX-Redirect", fmt.Sprintf("/dashboard/%s", teamId))
	w.WriteHeader(http.StatusOK)
}
