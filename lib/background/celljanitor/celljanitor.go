package celljanitor

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
	CellId string
}

// CheckInterval is how often we check in on the cell
const CheckInterval = 5 * time.Minute

// MessageHandler handles the message
type MessageHandler struct {
	q                   *background.QueueProducer[Message]
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
	return h.q.SendWithDelay(ctx, m, CheckInterval)
}

func (h MessageHandler) Handle(ctx context.Context, m Message) error {
	log := logger.FromContext(ctx).With(
		slog.String("cellId", m.CellId),
	)

	cell, err := h.cellStore.Get(m.CellId)
	if err != nil {
		return fmt.Errorf("error fetching cell: %v", err)
	}
	log = log.With(slog.String("cellType", string(cell.Type)))
	log.Info("janitoring cell")

	cellProvider := h.cellProviderForType(cell.Type)
	if cellProvider == nil {
		return fmt.Errorf("no cell provider found for cell type: %s", cell.Type)
	}

	if err := cellProvider.Janitor(ctx, cell.Id); err != nil {
		log.Error("error janitoring cell", slog.Any("error", err))
		return err
	}
	log.Info("cell janitored")
	return nil
}
