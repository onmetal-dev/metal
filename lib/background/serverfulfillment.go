package background

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/serverprovider"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/onmetal-dev/metal/lib/talosprovider"
	"github.com/siderolabs/talos/pkg/machinery/client"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
)

// ServerFulfillment message is sent as soon as a server is purchased. It takes care of the following:
// - creating the server object in the db
// - waiting for payment to be confirmed
// - ordering the server
// - waiting for the server to come online
// - setting up the server (installing Talos)
// If any of these steps involve waiting, then the logic will re-queue the message with a delay of 30s to check again later.
type ServerFulfillment struct {
	TeamId                  string
	UserId                  string
	OfferingId              string
	LocationId              string
	StripeCheckoutSessionId string
	// in the future these might be configurable on a per-fulfillment basis
	// e.g. to put a new server into an existing cell
	// or to create a new cell that has custom dns
	CellName  string // defaults to default cell
	DnsZoneId string

	StepServerId              string
	StepPaymentReceived       bool
	StepProviderTransactionId string
	StepProviderServerId      string
	StepServerOnline          bool
	StepServerInstalled       bool
	StepTalosOnline           bool
	StepServerAddedToCell     bool
}

// StepBuyServer buys the server and fills out the transaction ID
type StepBuyServer struct {
	TransactionId *string
}

const ServerFulfillmentCheckInterval = 30 * time.Second

type ServerFulfillmentHandler struct {
	q                     *QueueProducer[ServerFulfillment]
	teamStore             store.TeamStore
	userStore             store.UserStore
	serverStore           store.ServerStore
	serverOfferingStore   store.ServerOfferingStore
	cellStore             store.CellStore
	stripeCheckoutSession *session.Client
	serverProviderHetzner serverprovider.ServerProvider
	talosProviderHetzner  talosprovider.TalosProvider
	talosCellProvider     cellprovider.CellProvider
	sshKeyBase64          string
	sshKeyPassword        string
	sshKeyFingerprint     string
	logger                *slog.Logger
}

type ServerFulfillmentHandlerOption func(*ServerFulfillmentHandler) error

func WithQueueProducer(q *QueueProducer[ServerFulfillment]) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if q == nil {
			return errors.New("queue producer cannot be nil")
		}
		h.q = q
		return nil
	}
}

func WithTeamStore(teamStore store.TeamStore) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if teamStore == nil {
			return errors.New("team store cannot be nil")
		}
		h.teamStore = teamStore
		return nil
	}
}

func WithUserStore(userStore store.UserStore) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if userStore == nil {
			return errors.New("user store cannot be nil")
		}
		h.userStore = userStore
		return nil
	}
}

func WithServerStore(serverStore store.ServerStore) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if serverStore == nil {
			return errors.New("server store cannot be nil")
		}
		h.serverStore = serverStore
		return nil
	}
}

func WithServerOfferingStore(serverOfferingStore store.ServerOfferingStore) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if serverOfferingStore == nil {
			return errors.New("server offering store cannot be nil")
		}
		h.serverOfferingStore = serverOfferingStore
		return nil
	}
}

func WithCellStore(cellStore store.CellStore) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if cellStore == nil {
			return errors.New("cell store cannot be nil")
		}
		h.cellStore = cellStore
		return nil
	}
}

func WithStripeCheckoutSession(stripeCheckoutSession *session.Client) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if stripeCheckoutSession == nil {
			return errors.New("stripe checkout session cannot be nil")
		}
		h.stripeCheckoutSession = stripeCheckoutSession
		return nil
	}
}

func WithServerProviderHetzner(serverProviderHetzner serverprovider.ServerProvider) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if serverProviderHetzner == nil {
			return errors.New("server provider hetzner cannot be nil")
		}
		h.serverProviderHetzner = serverProviderHetzner
		return nil
	}
}

func WithTalosProviderHetzner(talosProviderHetzner *talosprovider.HetznerProvider) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if talosProviderHetzner == nil {
			return errors.New("talos provider hetzner cannot be nil")
		}
		h.talosProviderHetzner = talosProviderHetzner
		return nil
	}
}

func WithTalosCellProvider(talosCellProvider *cellprovider.TalosClusterCellProvider) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if talosCellProvider == nil {
			return errors.New("talos cell provider cannot be nil")
		}
		h.talosCellProvider = talosCellProvider
		return nil
	}
}

func WithSshKeyBase64(sshKeyBase64 string) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if sshKeyBase64 == "" {
			return errors.New("ssh key base64 cannot be empty")
		}
		h.sshKeyBase64 = sshKeyBase64
		return nil
	}
}

func WithSshKeyPassword(sshKeyPassword string) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if sshKeyPassword == "" {
			return errors.New("ssh key password cannot be empty")
		}
		h.sshKeyPassword = sshKeyPassword
		return nil
	}
}

func WithSshKeyFingerprint(sshKeyFingerprint string) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if sshKeyFingerprint == "" {
			return errors.New("ssh key fingerprint cannot be empty")
		}
		h.sshKeyFingerprint = sshKeyFingerprint
		return nil
	}
}

func WithLogger(logger *slog.Logger) ServerFulfillmentHandlerOption {
	return func(h *ServerFulfillmentHandler) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		h.logger = logger
		return nil
	}
}

func NewServerFulfillmentHandler(opts ...ServerFulfillmentHandlerOption) (*ServerFulfillmentHandler, error) {
	h := &ServerFulfillmentHandler{}
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
	if h.userStore == nil {
		errs = append(errs, "user store is required")
	}
	if h.serverStore == nil {
		errs = append(errs, "server store is required")
	}
	if h.serverOfferingStore == nil {
		errs = append(errs, "server offering store is required")
	}
	if h.cellStore == nil {
		errs = append(errs, "cell store is required")
	}
	if h.stripeCheckoutSession == nil {
		errs = append(errs, "stripe checkout session is required")
	}
	if h.serverProviderHetzner == nil {
		errs = append(errs, "server provider hetzner is required")
	}
	if h.talosProviderHetzner == nil {
		errs = append(errs, "talos provider hetzner is required")
	}
	if h.talosCellProvider == nil {
		errs = append(errs, "talos cell provider is required")
	}
	if h.sshKeyBase64 == "" {
		errs = append(errs, "ssh key base64 is required")
	}
	if h.sshKeyPassword == "" {
		errs = append(errs, "ssh key password is required")
	}
	if h.sshKeyFingerprint == "" {
		errs = append(errs, "ssh key fingerprint is required")
	}
	if h.logger == nil {
		errs = append(errs, "logger is required")
	}
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, ", "))
	}
	return h, nil
}

func (h ServerFulfillmentHandler) ReQueue(ctx context.Context, s ServerFulfillment) error {
	return h.q.SendWithDelay(ctx, s, ServerFulfillmentCheckInterval)
}

func (h ServerFulfillmentHandler) Handle(ctx context.Context, s ServerFulfillment) error {
	logger := h.logger.With(
		slog.String("teamId", s.TeamId),
		slog.String("userId", s.UserId),
		slog.String("offeringId", s.OfferingId),
		slog.String("location", s.LocationId),
	)
	offering, err := h.serverOfferingStore.GetServerOffering(s.OfferingId)
	if err != nil {
		logger.Error("Failed to get offering", "error", err)
		return err
	}
	if s.StepServerId == "" {
		logger.Info("Creating server in database")
		// sometimes we send the same fulfillment message after payment has cleared
		// but we want to re-create the server object in the db
		status := store.ServerStatusPendingPayment
		if s.StepPaymentReceived {
			status = store.ServerStatusPendingProvider
		}
		// sometimes we re-send the same fulfillment message w/ a server already
		// created for the user, in which case we should populate the provider
		// server id now
		var providerId *string
		if s.StepProviderServerId != "" {
			providerId = &s.StepProviderServerId
		}
		server, err := h.serverStore.Create(store.Server{
			TeamId:       s.TeamId,
			UserId:       s.UserId,
			ProviderSlug: string(offering.ProviderSlug),
			OfferingId:   s.OfferingId,
			LocationId:   s.LocationId,
			Status:       status,
			ProviderId:   providerId,
		})
		if err != nil {
			logger.Error("Failed to create server", "error", err)
			return err
		}
		s.StepServerId = server.Id
		logger.Info("Server created in database", "serverId", s.StepServerId)
	}
	logger = logger.With(slog.String("serverId", s.StepServerId))

	if !s.StepPaymentReceived {
		logger.Info("Checking payment status", "checkoutSessionId", s.StripeCheckoutSessionId)
		checkoutSession, err := h.stripeCheckoutSession.Get(s.StripeCheckoutSessionId, nil)
		if err != nil {
			logger.Error("Failed to get checkout session", "error", err)
			return err
		}
		if !(checkoutSession.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid ||
			checkoutSession.PaymentStatus == stripe.CheckoutSessionPaymentStatusNoPaymentRequired) {
			logger.Info("Payment not yet received, re-queueing", "paymentStatus", checkoutSession.PaymentStatus)
			return h.ReQueue(ctx, s)
		}
		if err := h.serverStore.UpdateServerStatus(s.StepServerId, store.ServerStatusPendingProvider); err != nil {
			logger.Error("Failed to update server status", "error", err)
			return err
		}
		s.StepPaymentReceived = true
		logger.Info("Payment received and server status updated")
	}

	provider, err := h.getServerProvider(string(offering.ProviderSlug))
	if err != nil {
		return err
	}

	if s.StepProviderTransactionId == "" {
		logger.Info("Ordering server from provider")
		transaction, err := provider.OrderServer(serverprovider.Order{
			OfferingId: s.OfferingId,
			LocationId: s.LocationId,
		})
		if err != nil {
			logger.Error("Failed to order server", "error", err)
			return err
		}
		s.StepProviderTransactionId = transaction.Id
		logger.Info("Server ordered from provider", "transactionId", s.StepProviderTransactionId)
	}
	logger = logger.With(slog.String("transactionId", s.StepProviderTransactionId))

	if s.StepProviderServerId == "" {
		logger.Info("Checking if server has a provider ID")
		tx, err := provider.GetTransaction(s.StepProviderTransactionId)
		if err != nil {
			logger.Error("Failed to get transaction", "error", err)
			return err
		}
		if tx.ServerId != "" {
			s.StepProviderServerId = tx.ServerId
			if err := h.serverStore.UpdateProviderId(s.StepServerId, tx.ServerId); err != nil {
				logger.Error("Failed to update server provider ID", "error", err)
				return err
			}
		} else {
			logger.Info("Server doesn't have a provider ID yet, re-queueing")
			return h.ReQueue(ctx, s)
		}
	}
	logger = logger.With(slog.String("providerServerId", s.StepProviderServerId))

	if !s.StepServerOnline {
		logger.Info("Checking if server is online")
		server, err := provider.GetServer(s.StepProviderServerId)
		if err != nil {
			logger.Error("Failed to get server status", "error", err)
			return err
		}
		if server.Ipv4 == "" {
			logger.Info("Server not yet online (no ipv4), re-queueing")
			return h.ReQueue(ctx, s)
		}
		if err := h.serverStore.UpdateServerPublicIpv4(s.StepServerId, server.Ipv4); err != nil {
			logger.Error("Failed to update server public ipv4", "error", err)
			return err
		}
		s.StepServerOnline = server.Status == serverprovider.ServerStatusRunning
		if !s.StepServerOnline {
			logger.Info("Server not yet online, re-queueing", "serverStatus", server.Status)
			return h.ReQueue(ctx, s)
		}
		logger.Info("Server is online")
	}

	if !s.StepServerInstalled {
		logger.Info("Installing Talos on server")
		talosProvider, err := h.getTalosProvider(string(offering.ProviderSlug))
		if err != nil {
			return err
		}
		server, err := provider.GetServer(s.StepProviderServerId)
		if err != nil {
			return err
		}
		if err := talosProvider.Install(ctx, talosprovider.Server{
			Id:                    server.Id,
			Ip:                    server.Ipv4,
			Username:              "root",
			SshKeyPrivateBase64:   h.sshKeyBase64,
			SshKeyPrivatePassword: h.sshKeyPassword,
			SshKeyFingerprint:     h.sshKeyFingerprint,
		}, talosprovider.WithTalosVersion("1.7.6"), talosprovider.WithArch("amd64")); err != nil {
			logger.Error("Failed to install Talos", "error", err)
			return err
		}
		s.StepServerInstalled = true
	}

	if !s.StepTalosOnline {
		logger.Info("Checking if Talos is online")
		server, err := provider.GetServer(s.StepProviderServerId)
		if err != nil {
			return err
		}
		c, err := client.New(ctx, client.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}), client.WithEndpoints(server.Ipv4))
		if err != nil {
			return err
		}
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if _, err := c.Disks(ctxWithTimeout); err != nil {
			logger.Info("Talos not yet online, re-queueing", slog.String("error", err.Error()))
			return h.ReQueue(ctx, s)
		}
		s.StepTalosOnline = true
	}

	if !s.StepServerAddedToCell {
		logger.Info("Adding server to cell")
		var c *store.Cell
		if cells, err := h.cellStore.GetForTeam(s.TeamId); err != nil {
			return err
		} else if len(cells) > 0 {
			for _, cell := range cells {
				if s.CellName != "" && cell.Name == s.CellName {
					c = &cell
					break
				} else if s.CellName == "" && cell.Name == "default" {
					c = &cell
				}
			}
		} else {
			created, err := h.cellStore.Create(store.Cell{
				Name:   "default",
				TeamId: s.TeamId,
			})
			if err != nil {
				return err
			}
			c = &created
		}

		// TODO: at this point we either have a
		// 1. freshly minted default cell that needs to be set up with talos
		// 2. a cell that already exists and is set up with talos
		// 3. (future) a cell that already exists and is not a talos cell
		if c.TalosCellData != nil {
			return fmt.Errorf("cell already has talos installed, TODO: add server to existing talos cluster")
		} else {
			// use taloscellprovider to create a new single-node talos cluster with this server
			team, err := h.teamStore.GetTeam(s.TeamId)
			if err != nil {
				return err
			}
			server, err := h.serverStore.Get(s.StepServerId)
			if err != nil {
				return err
			}
			if _, err := h.talosCellProvider.CreateCell(ctx, cellprovider.CreateCellOptions{
				Name:              c.Name,
				TeamId:            team.Id,
				TeamName:          team.Name,
				TeamAgePrivateKey: team.AgePrivateKey,
				DnsZoneId:         s.DnsZoneId,
				FirstServer:       server,
			}); err != nil {
				return err
			}
		}
	}

	logger.Info("Server fulfillment completed successfully")
	return nil
}

func (h ServerFulfillmentHandler) getServerProvider(providerSlug string) (serverprovider.ServerProvider, error) {
	switch providerSlug {
	case "hetzner":
		return h.serverProviderHetzner, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerSlug)
	}
}

func (h ServerFulfillmentHandler) getTalosProvider(providerSlug string) (talosprovider.TalosProvider, error) {
	switch providerSlug {
	case "hetzner":
		return h.talosProviderHetzner, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerSlug)
	}
}
