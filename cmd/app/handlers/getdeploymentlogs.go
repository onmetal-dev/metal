package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/form"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/samber/lo"
)

const MaxLogsBeforeUiGetsWeird = 10000

type GetDeploymentLogsHandler struct {
	teamStore           store.TeamStore
	deploymentStore     store.DeploymentStore
	cellProviderForType func(cellType store.CellType) cellprovider.CellProvider
}

func NewGetDeploymentLogsHandler(teamStore store.TeamStore, deploymentStore store.DeploymentStore, cellProviderForType func(cellType store.CellType) cellprovider.CellProvider) *GetDeploymentLogsHandler {
	return &GetDeploymentLogsHandler{
		teamStore:           teamStore,
		deploymentStore:     deploymentStore,
		cellProviderForType: cellProviderForType,
	}
}

func (h *GetDeploymentLogsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	appId := chi.URLParam(r, "appId")
	envId := chi.URLParam(r, "envId")
	deploymentIdStr := chi.URLParam(r, "deploymentId")
	user := middleware.GetUser(ctx)
	team, teams := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	deploymentId, err := strconv.Atoi(deploymentIdStr)
	if err != nil {
		http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
		return
	}

	deployment, err := h.deploymentStore.Get(appId, envId, uint(deploymentId))
	if err != nil {
		http.Error(w, "error fetching deployment", http.StatusInternalServerError)
		return
	}

	if deployment.TeamId != teamId {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// the defaults
	defaultFormData := templates.LogsFormData{
		Since: "15m",
	}

	getLogs := func(fd templates.LogsFormData) ([]cellprovider.LogEntry, error) {
		var duration = 15 * time.Minute
		if fd.Since != "" {
			var err error
			duration, err = time.ParseDuration(fd.Since)
			if err != nil {
				return nil, fmt.Errorf("error parsing duration: %s", err)
			}
		}
		logs := []cellprovider.LogEntry{}
		for _, cell := range deployment.Cells {
			les, err := h.cellProviderForType(cell.Type).DeploymentLogs(ctx, cell.Id, &deployment, cellprovider.WithSince(duration))
			if err != nil {
				return nil, fmt.Errorf("error fetching deployment logs: %s", err)
			}
			logs = append(logs, les...)
		}
		sort.Slice(logs, func(i, j int) bool {
			return logs[i].Timestamp.After(logs[j].Timestamp)
		})
		// searching logs is my passion
		if fd.Query != "" {
			logs = lo.Filter(logs, func(log cellprovider.LogEntry, _ int) bool {
				return strings.Contains(strings.ToLower(log.Message), strings.ToLower(fd.Query))
			})
		}
		return logs, nil
	}

	if r.Method == "POST" {
		// posting the form. pull in form data
		var f templates.LogsFormData
		fieldErrs, err := form.Decode(&f, r)
		if fieldErrs.NotNil() || err != nil {
			// send back the form html w/ errors
			if err := templates.LogsForm(r.URL.Path, f, fieldErrs, err, []cellprovider.LogEntry{}, false, "").Render(ctx, w); err != nil {
				http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
			}
			return
		}
		logs, err := getLogs(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		warning := ""
		if len(logs) > MaxLogsBeforeUiGetsWeird {
			logs = logs[:MaxLogsBeforeUiGetsWeird]
			warning = fmt.Sprintf("search limit exceeded. only showing the most recent %d logs", MaxLogsBeforeUiGetsWeird)
		}
		if err := templates.LogsForm(r.URL.Path, f, fieldErrs, err, logs, false, warning).Render(ctx, w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		err = templates.DashboardLayout(templates.DashboardState{
			User:          *user,
			Teams:         teams,
			ActiveTeam:    *team,
			ActiveTabName: templates.TabNameHome,
		}, templates.DeploymentLogs(deployment, r.URL.Path, defaultFormData, form.FieldErrors{}, nil, []cellprovider.LogEntry{})).Render(ctx, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
