package deployment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"github.com/onmetal-dev/metal/lib/background"
	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/store"
)

// Message contains the deployment ID to manage and monitor
type Message struct {
	DeploymentId uint
	AppId        string
	EnvId        string
}

const DeploymentCheckInterval = 5 * time.Second

// MessageHandler handles the deployment message
type MessageHandler struct {
	q                   *background.QueueProducer[Message]
	deploymentStore     store.DeploymentStore
	cellStore           store.CellStore
	cellProviderForType func(cellType store.CellType) cellprovider.CellProvider
}

type Option func(*MessageHandler) error

func WithQueueProducer(q *background.QueueProducer[Message]) Option {
	return func(h *MessageHandler) error {
		if q == nil {
			return errors.New("queue producer cannot be nil")
		}
		h.q = q
		return nil
	}
}

func WithDeploymentStore(deploymentStore store.DeploymentStore) Option {
	return func(h *MessageHandler) error {
		if deploymentStore == nil {
			return errors.New("deployment store cannot be nil")
		}
		h.deploymentStore = deploymentStore
		return nil
	}
}

func WithCellProviderForType(fn func(cellType store.CellType) cellprovider.CellProvider) Option {
	return func(h *MessageHandler) error {
		if fn == nil {
			return errors.New("cell provider function cannot be nil")
		}
		h.cellProviderForType = fn
		return nil
	}
}

func WithCellStore(cellStore store.CellStore) Option {
	return func(h *MessageHandler) error {
		if cellStore == nil {
			return errors.New("cell store cannot be nil")
		}
		h.cellStore = cellStore
		return nil
	}
}

func NewMessageHandler(opts ...Option) (*MessageHandler, error) {
	h := &MessageHandler{}
	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}
	var errs []string
	if h.q == nil {
		errs = append(errs, "queue producer is required")
	}
	if h.deploymentStore == nil {
		errs = append(errs, "deployment store is required")
	}
	if h.cellStore == nil {
		errs = append(errs, "cell store is required")
	}
	if h.cellProviderForType == nil {
		errs = append(errs, "cell provider for type function is required")
	}
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ", "))
	}
	return h, nil
}

func (h MessageHandler) ReQueue(ctx context.Context, m Message) error {
	return h.q.SendWithDelay(ctx, m, DeploymentCheckInterval)
}

func (h MessageHandler) Handle(ctx context.Context, m Message) error {
	log := logger.FromContext(ctx).With(
		slog.Int("deploymentID", int(m.DeploymentId)),
		slog.String("appID", m.AppId),
		slog.String("envID", m.EnvId),
	)

	deployment, err := h.deploymentStore.Get(m.AppId, m.EnvId, m.DeploymentId)
	if err != nil {
		return fmt.Errorf("error fetching deployment: %v", err)
	}

	if deployment.Status == store.DeploymentStatusRunning || deployment.Status == store.DeploymentStatusFailed {
		log.Info("Deployment already in final state, no action needed")
		return nil
	}

	if len(deployment.Cells) == 0 {
		return fmt.Errorf("no cells associated with deployment")
	}

	cell, err := h.cellStore.Get(deployment.Cells[0].Id)
	if err != nil {
		return fmt.Errorf("error fetching cell: %v", err)
	}

	cellProvider := h.cellProviderForType(cell.Type)
	if cellProvider == nil {
		return fmt.Errorf("no cell provider found for cell type: %s", cell.Type)
	}

	result, err := cellProvider.AdvanceDeployment(ctx, cell.Id, &deployment)
	if err != nil {
		log.Error("Error advancing deployment", slog.Any("error", err))
		if updateErr := h.deploymentStore.UpdateDeploymentStatus(m.AppId, m.EnvId, m.DeploymentId, store.DeploymentStatusFailed, err.Error()); updateErr != nil {
			log.Error("Error updating deployment status", slog.Any("error", updateErr))
		}
		return err
	}

	if err := h.deploymentStore.UpdateDeploymentStatus(m.AppId, m.EnvId, m.DeploymentId, result.Status, result.StatusReason); err != nil {
		log.Error("Error updating deployment status", slog.Any("error", err))
		return err
	}

	if result.StatusReason != "" {
		log.Info("Deployment status update", slog.String("status", string(result.Status)), slog.String("reason", result.StatusReason))
	} else {
		log.Info("Deployment status update", slog.String("status", string(result.Status)))
	}

	if result.Status == store.DeploymentStatusDeploying {
		log.Info("Deployment still in progress, requeueing")
		return h.ReQueue(ctx, m)
	}

	return nil
}
