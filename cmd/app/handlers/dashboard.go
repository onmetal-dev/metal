package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/cmd/app/urls"
	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/logger"
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

type SseEvent struct {
	EventName string
	Data      string
}

func (e *SseEvent) String() string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", e.EventName, e.Data)
}

func serverStatToEvents(ctx context.Context, stats cellprovider.ServerStats) []*SseEvent {
	events := make([]*SseEvent, 2)
	{
		var buffer bytes.Buffer
		if err := templates.ServerStatsCpu(&stats).Render(ctx, &buffer); err != nil {
			return nil
		}
		events[0] = &SseEvent{
			EventName: templates.ServerStatsCpuSseEventName(stats.ServerId),
			Data:      buffer.String(),
		}
	}
	{
		var buffer bytes.Buffer
		if err := templates.ServerStatsMem(&stats).Render(ctx, &buffer); err != nil {
			return nil
		}
		events[1] = &SseEvent{
			EventName: templates.ServerStatsMemSseEventName(stats.ServerId),
			Data:      buffer.String(),
		}
	}
	return events
}

func (h *DashboardHandler) ServeHTTPSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, _ := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}

	cells, err := h.cellStore.GetForTeam(ctx, teamId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for sending events
	events := make(chan *SseEvent)

	// Create a context that cancels when the client disconnects
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-r.Context().Done()
		logger.FromContext(r.Context()).Info("client disconnected")
		cancel()
	}()

	// Start a goroutine for each cell to stream server stats
	var wg sync.WaitGroup
	for _, cell := range cells {
		wg.Add(1)
		go func(cell store.Cell) {
			defer wg.Done()
			statsChan := h.cellProviderForType(cell.Type).ServerStatsStream(ctx, cell.Id, 5*time.Second)
			for result := range statsChan {
				if result.Error != nil {
					logger.FromContext(ctx).Error("error fetching server stats", "error", result.Error, "cellId", cell.Id)
					continue
				}
				for _, stat := range result.Stats {
					for _, event := range serverStatToEvents(ctx, stat) {
						select {
						case events <- event:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}(cell)
	}

	// Start a goroutine to close the events channel when all cell streams are done
	go func() {
		wg.Wait()
		close(events)
	}()

	// Send events to the client
	for event := range events {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := fmt.Fprint(w, event.String())
			if err != nil {
				logger.FromContext(ctx).Error("error writing event to response", "error", err)
				return
			}
			w.(http.Flusher).Flush()
		}
	}
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	envName := chi.URLParam(r, "envName")
	user := middleware.GetUser(ctx)
	team, userTeams := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	var (
		servers     []store.Server
		cells       []store.Cell
		deployments []store.Deployment
		apps        []store.App
		envs        []store.Env
		activeEnv   store.Env
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
		deployments, err = h.deploymentStore.GetForTeam(ctx, teamId)
		return err
	})

	g.Go(func() error {
		var err error
		apps, err = h.appStore.GetForTeam(ctx, teamId)
		return err
	})

	g.Go(func() error {
		var err error
		envs, err = h.deploymentStore.GetEnvsForTeam(teamId)
		if err != nil {
			return err
		}
		for _, env := range envs {
			if env.Name == envName {
				activeEnv = env
			}
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if activeEnv.Id == "" {
		if envName == urls.DefaultEnvSentinel {
			// redirect to the first env
			envName = envs[0].Name
			http.Redirect(w, r, urls.Home{TeamId: teamId, EnvName: envName}.Render(), http.StatusTemporaryRedirect)
			return
		}
		http.Error(w, "env not found", http.StatusNotFound)
		return
	}

	if err := templates.DashboardLayout(templates.DashboardState{
		User:          *user,
		Teams:         userTeams,
		ActiveTeam:    *team,
		Envs:          envs,
		ActiveEnv:     &activeEnv,
		ActiveTabName: templates.TabNameHome,
		AdditionalScripts: []templates.ScriptTag{
			templates.ScriptTag{
				Src: "/static/script/sse.js",
			},
		},
	}, templates.DashboardHome(teamId, envName, servers, cells, deployments, apps)).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}
