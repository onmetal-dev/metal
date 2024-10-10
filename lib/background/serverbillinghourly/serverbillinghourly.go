package serverbillinghourly

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"github.com/onmetal-dev/metal/lib/background"
	"github.com/onmetal-dev/metal/lib/billing"
	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/billing/meterevent"
)

// Message bills for a server hourly, sending a meter event to stripe w/ the correct usage data since the last time we sent a meter event to stripe.
type Message struct {
	TeamId           string
	OfferingId       string
	LocationId       string
	StripeCustomerId string
	ServerId         string
}

const MessageRequeueInterval = 1 * time.Hour

// MessageHandler handles the message
type MessageHandler struct {
	q                *background.QueueProducer[Message]
	teamStore        store.TeamStore
	offeringStore    store.ServerOfferingStore
	serverStore      store.ServerStore
	stripeMeterEvent *meterevent.Client
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

func WithTeamStore(teamStore store.TeamStore) Option {
	return func(h *MessageHandler) error {
		if teamStore == nil {
			return errors.New("team store cannot be nil")
		}
		h.teamStore = teamStore
		return nil
	}
}

func WithServerOfferingStore(offeringStore store.ServerOfferingStore) Option {
	return func(h *MessageHandler) error {
		if offeringStore == nil {
			return errors.New("server offering store cannot be nil")
		}
		h.offeringStore = offeringStore
		return nil
	}
}

func WithServerStore(serverStore store.ServerStore) Option {
	return func(h *MessageHandler) error {
		if serverStore == nil {
			return errors.New("server store cannot be nil")
		}
		h.serverStore = serverStore
		return nil
	}
}

func WithStripeMeterEvent(stripeMeterEvent *meterevent.Client) Option {
	return func(h *MessageHandler) error {
		if stripeMeterEvent == nil {
			return errors.New("stripe meter event client cannot be nil")
		}
		h.stripeMeterEvent = stripeMeterEvent
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
	if h.teamStore == nil {
		errs = append(errs, "team store is required")
	}
	if h.offeringStore == nil {
		errs = append(errs, "server offering store is required")
	}
	if h.serverStore == nil {
		errs = append(errs, "server store is required")
	}
	if h.stripeMeterEvent == nil {
		errs = append(errs, "stripe meter event client is required")
	}
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ", "))
	}
	return h, nil
}

func (h MessageHandler) ReQueue(ctx context.Context, m Message) error {
	return h.q.SendWithDelay(ctx, m, MessageRequeueInterval)
}

func (h MessageHandler) Handle(ctx context.Context, m Message) error {
	log := logger.FromContext(ctx).With(
		slog.String("teamId", m.TeamId),
		slog.String("offeringId", m.OfferingId),
		slog.String("location", m.LocationId),
		slog.String("serverId", m.ServerId),
		slog.String("stripeCustomerId", m.StripeCustomerId),
	)
	offering, err := h.offeringStore.GetServerOffering(m.OfferingId)
	if err != nil {
		return err
	}
	eventName := billing.UsageHourMeterEventName(*offering, m.LocationId)

	server, err := h.serverStore.Get(m.ServerId)
	if err != nil {
		return err
	}

	now := time.Now()
	var hoursToBill int
	if server.BillingStripeUsageBasedHourly == nil {
		hoursToBill = int(now.Sub(server.CreatedAt).Hours())
	} else {
		hoursToBill = int(now.Sub(server.BillingStripeUsageBasedHourly.LastEventSent.Time).Hours())
	}

	newBillingMeta := store.ServerBillingStripeUsageBasedHourly{
		ServerId:      server.Id,
		LastEventSent: sql.NullTime{Time: now, Valid: true},
		EventName:     eventName,
	}

	// set up a unique event ID that ensures we only send one event per hour per server
	unixTimestampRoundedDownToNearestHour := time.Now().Truncate(time.Hour).Unix()
	uniqueEventId := fmt.Sprintf("%s-%d", server.Id, unixTimestampRoundedDownToNearestHour)

	log.Info("hourly server billing event", "hoursToBill", hoursToBill, "uniqueEventId", uniqueEventId)
	if _, err := h.stripeMeterEvent.New(&stripe.BillingMeterEventParams{
		EventName: stripe.String(eventName),
		Payload: map[string]string{
			"value":              fmt.Sprintf("%d", hoursToBill),
			"stripe_customer_id": m.StripeCustomerId,
		},
		Identifier: stripe.String(uniqueEventId),
	}); err != nil {
		if strings.Contains(err.Error(), "event already exists with identifier") {
			log.Info("hourly server billing event already sent", "hoursToBill", hoursToBill, "uniqueEventId", uniqueEventId)
		} else {
			return err
		}
	}

	if err := h.serverStore.UpdateServerBillingStripeUsageBasedHourly(server.Id, &newBillingMeta); err != nil {
		return err
	}

	log.Info("hourly server billing completed successfully")
	return h.ReQueue(ctx, m)
}
