package talosprovider

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/floshodan/hrobot-go/hrobot"
)

type HetznerProvider struct {
	client *hrobot.Client
	logger *slog.Logger
}

type HetznerProviderOption func(*HetznerProvider)

func WithClient(client *hrobot.Client) HetznerProviderOption {
	return func(h *HetznerProvider) {
		h.client = client
	}
}

func WithLogger(logger *slog.Logger) HetznerProviderOption {
	return func(h *HetznerProvider) {
		h.logger = logger
	}
}

func NewHetznerProvider(opts ...HetznerProviderOption) (*HetznerProvider, error) {
	h := &HetznerProvider{}
	for _, opt := range opts {
		opt(h)
	}
	if h.client == nil {
		return nil, fmt.Errorf("client is required")
	}
	return h, nil
}

var _ TalosProvider = &HetznerProvider{}

func (h *HetznerProvider) Install(ctx context.Context, server Server, opts ...InstallOption) error {
	logger := h.logger.With(slog.String("serverId", server.Id), slog.String("serverIp", server.Ip))
	options := &InstallOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if err := options.Validate(); err != nil {
		return err
	}

	if _, _, err := h.client.Boot.ActivateRescue(ctx, server.Id, &hrobot.RescueOpts{
		OS:             "linux",
		Authorized_Key: server.SshKeyFingerprint,
		Keyboard:       "us",
	}); err != nil {
		return fmt.Errorf("error putting hetzner server into rescue mode: %v", err)
	}
	if _, _, err := h.client.Reset.ExecuteReset(ctx, server.Id, "hw"); err != nil {
		return fmt.Errorf("error resetting hetzner server: %v", err)
	}
	// check every 5s for port 22 to be open on server IP (hardcoded to 95.217.100.228)
	logger.Info("waiting for server to come up in rescue mode")
	for {
		time.Sleep(5 * time.Second)
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:22", server.Ip), 5*time.Second)
		if err == nil {
			conn.Close()
			break
		}
	}
	logger.Info("server is up and running")

	if err := StopRaids(server); err != nil {
		return fmt.Errorf("failed to stop raids: %v", err)
	}
	logger.Info("stopped raids")
	if err := SetupAndWipeFilesystem(server); err != nil {
		return fmt.Errorf("failed to setup and wipe filesystem: %v", err)
	}
	logger.Info("setup and wiped filesystem")
	if err := DownloadAndInstallTalos(server, options.TalosVersion, "metal", options.Arch); err != nil {
		return fmt.Errorf("failed to download and install talos: %v", err)
	}
	logger.Info("downloaded and installed talos")
	if _, err := createAndRunCommand(server, "reboot"); err != nil {
		return fmt.Errorf("failed to reboot server: %v", err)
	}
	logger.Info("rebooted server")
	return nil
}
