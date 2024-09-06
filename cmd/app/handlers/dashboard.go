package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
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

	getStats := func() ([]cellprovider.ServerStats, error) {
		var stats []cellprovider.ServerStats
		for _, cell := range cells {
			cellStats, err := h.cellProviderForType(cell.Type).ServerStats(ctx, cell.Id)
			if err != nil {
				return nil, fmt.Errorf("error fetching server stats: %v", err)
			}
			stats = append(stats, cellStats...)
		}
		return stats, nil
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for sending events
	events := make(chan *SseEvent)

	// Close the channel when the client disconnects
	go func() {
		<-r.Context().Done()
		logger.FromContext(r.Context()).Info("client disconnected")
		close(events)
	}()

	// Send an event every 5 seconds
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				stats, err := getStats()
				if err != nil {
					logger.FromContext(r.Context()).Error("error fetching server stats", "error", err)
					continue
				}
				for _, stat := range stats {
					for _, event := range serverStatToEvents(ctx, stat) {
						events <- event
					}
				}
			case <-r.Context().Done():
				return
			}
		}
	}()

	// Send events to the client
	for {
		select {
		case event := <-events:
			fmt.Fprint(w, event.String())
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
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
	}, templates.DashboardHome(teamId, servers, cells, deployments, apps)).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}
