package cellprovider

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"filippo.io/age"
	"github.com/asaskevich/govalidator"
	thconfig "github.com/budimanjojo/talhelper/v3/pkg/config"
	"github.com/budimanjojo/talhelper/v3/pkg/generate"
	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/aes"
	keysource "github.com/getsops/sops/v3/age"
	"github.com/getsops/sops/v3/cmd/sops/common"
	"github.com/getsops/sops/v3/keys"
	"github.com/getsops/sops/v3/keyservice"
	sopsyaml "github.com/getsops/sops/v3/stores/yaml"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-yaml/yaml"
	"github.com/mholt/archiver/v4"
	"github.com/onmetal-dev/metal/lib/dnsprovider"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/samber/lo"
	"github.com/siderolabs/talos/cmd/talosctl/pkg/talos/helpers"
	"github.com/siderolabs/talos/pkg/machinery/api/machine"
	"github.com/siderolabs/talos/pkg/machinery/api/storage"
	"github.com/siderolabs/talos/pkg/machinery/client"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
	"github.com/siderolabs/talos/pkg/machinery/config"
	"github.com/siderolabs/talos/pkg/machinery/config/generate/secrets"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/ptr"
)

// TalosClusterCellProvider creates a talos k8s cluster using cloudflare for DNS
type TalosClusterCellProvider struct {
	dnsProvider    dnsprovider.DNSProvider
	cellStore      store.CellStore
	serverStore    store.ServerStore
	tmpDirRoot     string
	logger         *slog.Logger
	tracerProvider *trace.TracerProvider
}

var _ CellProvider = &TalosClusterCellProvider{}

type TalosClusterCellProviderOption func(*TalosClusterCellProvider)

func WithDnsProvider(dnsProvider dnsprovider.DNSProvider) TalosClusterCellProviderOption {
	return func(p *TalosClusterCellProvider) {
		p.dnsProvider = dnsProvider
	}
}

func WithCellStore(cellStore store.CellStore) TalosClusterCellProviderOption {
	return func(p *TalosClusterCellProvider) {
		p.cellStore = cellStore
	}
}

func WithServerStore(serverStore store.ServerStore) TalosClusterCellProviderOption {
	return func(p *TalosClusterCellProvider) {
		p.serverStore = serverStore
	}
}

func WithTmpDirRoot(tmpDirRoot string) TalosClusterCellProviderOption {
	return func(p *TalosClusterCellProvider) {
		p.tmpDirRoot = tmpDirRoot
	}
}

func WithLogger(logger *slog.Logger) TalosClusterCellProviderOption {
	return func(p *TalosClusterCellProvider) {
		p.logger = logger
	}
}

func WithTracerProvider(tp *trace.TracerProvider) TalosClusterCellProviderOption {
	return func(p *TalosClusterCellProvider) {
		p.tracerProvider = tp
	}
}

func NewTalosClusterCellProvider(opts ...TalosClusterCellProviderOption) (*TalosClusterCellProvider, error) {
	provider := &TalosClusterCellProvider{}
	for _, opt := range opts {
		opt(provider)
	}
	var errs []error
	if provider.dnsProvider == nil {
		errs = append(errs, fmt.Errorf("must provide a valid DNS provider"))
	}
	if provider.cellStore == nil {
		errs = append(errs, fmt.Errorf("must provide a valid cell store"))
	}
	if provider.serverStore == nil {
		errs = append(errs, fmt.Errorf("must provide a valid server store"))
	}
	if provider.tmpDirRoot == "" {
		errs = append(errs, fmt.Errorf("must provide a valid tmpDirRoot"))
	}
	if provider.logger == nil {
		errs = append(errs, fmt.Errorf("must provide a valid logger"))
	}
	if provider.tracerProvider == nil {
		errs = append(errs, fmt.Errorf("must provide a valid tracer provider"))
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("errors: %v", errs)
	}
	return provider, nil
}

var sopsEnvMutex sync.Mutex

// CreateCell for a talos cluster creates a single-node talos cluster.
func (p *TalosClusterCellProvider) CreateCell(ctx context.Context, opts CreateCellOptions) (*store.Cell, error) {
	talosVersion := "1.7.6"
	if _, err := govalidator.ValidateStruct(opts); err != nil {
		return nil, fmt.Errorf("error validating createcell options: %v", err)
	}

	if err := p.dnsProvider.FindOrCreateARecord(ctx, opts.DnsZoneId, opts.FirstServer.Id, *opts.FirstServer.PublicIpv4); err != nil {
		return nil, err
	}
	domain, err := p.dnsProvider.Domain()
	if err != nil {
		return nil, err
	}

	// figure out install disk
	c, err := client.New(ctx,
		client.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
		client.WithEndpoints(*opts.FirstServer.PublicIpv4),
		client.WithGRPCDialOptions(grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithTracerProvider(p.tracerProvider)))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create talos client: %w", err)
	}
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	disksResp, err := c.Disks(ctxWithTimeout)
	if err != nil {
		return nil, fmt.Errorf("error getting disks: %w", err)
	}
	systemDisk, ok := lo.Find(disksResp.Messages[0].Disks, func(d *storage.Disk) bool {
		return d.SystemDisk
	})
	if !ok {
		return nil, fmt.Errorf("system disk not found")
	}

	// craft the single-node talos config
	thConfig := thconfig.TalhelperConfig{
		ClusterName:  opts.Name,
		TalosVersion: talosVersion,
		// todo: this works for single-node control plane setup, but once we're at > 1 node we'll need to use a setup like
		// <cell id>.cp.onmetal.run with an A record for each control plane node
		Endpoint:                 fmt.Sprintf("https://%s.%s:6443", opts.FirstServer.Id, domain),
		AllowSchedulingOnMasters: true,
		KubernetesVersion:        "v1.30.3",
		Patches: []string{
			`- op: add
  path: /cluster/discovery/enabled
  value: true`,
			`- op: replace
  path: /machine/network/kubespan
  value:
    enabled: true`,
			`- op: add
  path: /machine/kubelet/extraArgs
  value:
    rotate-server-certificates: "true"`,
		},
		Nodes: []thconfig.Node{
			{
				Hostname:     "cp-1",
				IPAddress:    *opts.FirstServer.PublicIpv4,
				ControlPlane: true,
				InstallDisk:  systemDisk.DeviceName,
				NodeConfigs: thconfig.NodeConfigs{
					NodeLabels: generateNodeLabels(NodeLabelInfo{
						ServerId:     opts.FirstServer.Id,
						CellName:     opts.Name,
						ProviderSlug: opts.FirstServer.ProviderSlug,
					}),
				},
			},
		},
	}
	errs, warnings := thConfig.Validate()
	if len(warnings) > 0 {
		for _, warning := range warnings {
			p.logger.Info(warning.Message)
		}
	}
	if len(errs) > 0 {
		errMsg := ""
		for _, err := range errs {
			errMsg += fmt.Sprintf("%s\n", err.Message)
		}
		return nil, fmt.Errorf("error validating talhelper config: %s", errMsg)
	}
	// do the equivalent of `talhelper gensecret`
	version, err := config.ParseContractFromVersion(talosVersion)
	if err != nil {
		return nil, fmt.Errorf("error parsing version contract: %w", err)
	}
	secretsBundle, err := secrets.NewBundle(secrets.NewClock(), version)
	if err != nil {
		return nil, fmt.Errorf("error creating secrets bundle: %w", err)
	}
	bs, err := yaml.Marshal(secretsBundle)
	if err != nil {
		return nil, fmt.Errorf("error marshalling secrets bundle: %w", err)
	}
	secretsBundleStr := string(bs)

	// do the equivalent of `sops -e -i talsecret.sops.yaml`
	encryptedSecretsBundleStr, err := encryptYaml(secretsBundleStr, opts.TeamAgePrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error encrypting secrets bundle: %w", err)
	}

	// set up a git repo that will be used to hold cluster config
	tempDir, err := os.MkdirTemp(p.tmpDirRoot, "cluster")
	if err != nil {
		return nil, fmt.Errorf("error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	repo, err := git.PlainInit(tempDir, false)
	if err != nil {
		return nil, fmt.Errorf("error initializing git repository: %v", err)
	}

	// Add the encrypted cluster secrets file to the directory along with talconfig.yaml, then do the equivalent of `talhelper genconfig`
	secretFilePath := filepath.Join(tempDir, "talsecret.sops.yaml")
	if err := os.WriteFile(secretFilePath, []byte(encryptedSecretsBundleStr), 0644); err != nil {
		return nil, fmt.Errorf("error writing file: %v", err)
	}
	talconfigFilePath := filepath.Join(tempDir, "talconfig.yaml")
	thConfigBs, err := yaml.Marshal(thConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshalling talhelper config: %w", err)
	}
	thConfigStr := string(thConfigBs)
	if err := os.WriteFile(talconfigFilePath, []byte(thConfigStr), 0644); err != nil {
		return nil, fmt.Errorf("error writing file: %v", err)
	}
	gitignoreFilePath := filepath.Join(tempDir, ".gitignore")
	if err := os.WriteFile(gitignoreFilePath, []byte("clusterconfig/"), 0644); err != nil {
		return nil, fmt.Errorf("error writing file: %v", err)
	}
	// this is hacky but there is no other way afaik
	sopsEnvMutex.Lock()
	os.Setenv("SOPS_AGE_KEY", opts.TeamAgePrivateKey)
	if err := generate.GenerateConfig(&thConfig, false, filepath.Join(tempDir, "clusterconfig"), secretFilePath, "metal", false); err != nil {
		os.Unsetenv("SOPS_AGE_KEY")
		sopsEnvMutex.Unlock()
		return nil, fmt.Errorf("error generating config: %v", err)
	}
	os.Unsetenv("SOPS_AGE_KEY")
	sopsEnvMutex.Unlock()

	// apply!
	cpConfigFilePath := filepath.Join(tempDir, "clusterconfig", fmt.Sprintf("%s-cp-1.yaml", opts.Name))
	cfgBytes, err := os.ReadFile(cpConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration from %s: %w", cpConfigFilePath, err)
	} else if len(cfgBytes) < 1 {
		return nil, errors.New("no configuration data read")
	}
	resp, err := c.ApplyConfiguration(ctx, &machine.ApplyConfigurationRequest{
		Data: cfgBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("error applying new configuration: %s", err)
	}

	helpers.PrintApplyResults(resp)

	// wait for port 50000 on the IP to be ready
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:50000", *opts.FirstServer.PublicIpv4), 5*time.Second)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(1 * time.Second)
	}

	// bootstrap with a new, secure client
	c, err = client.New(ctx,
		client.WithConfigFromFile(filepath.Join(tempDir, "clusterconfig", "talosconfig")),
		client.WithContextName(opts.Name),
		client.WithGRPCDialOptions(grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithTracerProvider(p.tracerProvider)))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create talos client from config: %w", err)
	}
	if err := c.Bootstrap(ctx, &machine.BootstrapRequest{}); err != nil {
		return nil, fmt.Errorf("error bootstrapping: %s", err)
	}

	// Commit the changes
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("error getting worktree: %v", err)
	}
	if err := worktree.AddWithOptions(&git.AddOptions{
		All: true,
	}); err != nil {
		return nil, fmt.Errorf("error adding file to repository: %v", err)
	}
	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Metal",
			Email: "automated@onmetal.dev",
			When:  time.Now(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error committing changes: %v", err)
	}

	// todo: wait for bootstrap to complete successfully (or make this another method on cellprovider to protect us from long-running fns)

	// Create an archive
	archiveTempDir, err := os.MkdirTemp(p.tmpDirRoot, "archive")
	if err != nil {
		return nil, fmt.Errorf("error creating archive temp directory: %v", err)
	}
	defer os.RemoveAll(archiveTempDir)
	zipFilePath := filepath.Join(archiveTempDir, "repository.zip")
	if err = archiveRepository(tempDir, zipFilePath); err != nil {
		return nil, fmt.Errorf("error archiving repository: %v", err)
	}
	zipBytes, err := os.ReadFile(zipFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading zip file: %v", err)
	}

	// update cell in the database to contain the new server and to have a new talosconfig
	talosConfig, err := os.ReadFile(filepath.Join(tempDir, "clusterconfig", "talosconfig"))
	if err != nil {
		return nil, fmt.Errorf("error reading talosconfig file: %v", err)
	}
	kubecfg, err := c.Kubeconfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting kubeconfig: %v", err)
	}

	cell, err := p.cellStore.Create(store.Cell{
		Name:    opts.Name,
		Type:    store.CellTypeTalos,
		TeamId:  opts.TeamId,
		Servers: []store.Server{opts.FirstServer},
		TalosCellData: &store.TalosCellData{
			Talosconfig: string(talosConfig),
			Kubecfg:     string(kubecfg),
			Config:      zipBytes,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating cell: %v", err)
	}
	return &cell, nil
}

type clientSetup struct {
	k8sClient   *kubernetes.Clientset
	talosClient *client.Client
	nodeIps     []string
}

func (p *TalosClusterCellProvider) setupClients(ctx context.Context, cellId string) (*clientSetup, error) {
	cell, err := p.cellStore.Get(cellId)
	if err != nil {
		return nil, fmt.Errorf("error getting cell: %v", err)
	}

	k8sClient, err := initializeK8sClient(cell.TalosCellData.Kubecfg, p.tracerProvider)
	if err != nil {
		return nil, fmt.Errorf("error initializing k8s client: %v", err)
	}

	talosClientConfig, err := clientconfig.FromString(cell.TalosCellData.Talosconfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing talosconfig: %v", err)
	}

	talosClient, err := client.New(ctx,
		client.WithConfig(talosClientConfig),
		client.WithGRPCDialOptions(grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithTracerProvider(p.tracerProvider)))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create talos client from config: %w", err)
	}

	nodeIps := talosClientConfig.Contexts[talosClientConfig.Context].Nodes

	return &clientSetup{
		k8sClient:   k8sClient,
		talosClient: talosClient,
		nodeIps:     nodeIps,
	}, nil
}

// getServerStatsWithClients retrieves server statistics using pre-initialized clients
func (p *TalosClusterCellProvider) getServerStatsWithClients(ctx context.Context, k8sClient *kubernetes.Clientset, talosClient *client.Client, nodeIps []string) ([]ServerStats, error) {
	var wg sync.WaitGroup
	var nodeInfo map[string]NodeInfo
	var nodeInfoErr error
	var systemStatsResp *machine.SystemStatResponse
	var systemStatsErr error
	var memoryResp *machine.MemoryResponse
	var memoryErr error

	wg.Add(3)

	go func() {
		defer wg.Done()
		nodeInfo, nodeInfoErr = getNodeIpv4ToLabels(ctx, k8sClient)
	}()

	go func() {
		defer wg.Done()
		systemStatsResp, systemStatsErr = talosClient.MachineClient.SystemStat(ctx, &emptypb.Empty{})
	}()

	go func() {
		defer wg.Done()
		memoryResp, memoryErr = talosClient.Memory(ctx)
	}()

	wg.Wait()

	if nodeInfoErr != nil {
		return nil, fmt.Errorf("error getting node info: %v", nodeInfoErr)
	}

	if systemStatsErr != nil {
		return nil, fmt.Errorf("error getting system stats: %v", systemStatsErr)
	}

	if memoryErr != nil {
		return nil, fmt.Errorf("error getting memory stats: %v", memoryErr)
	}

	result := make([]ServerStats, len(nodeIps))
	for i, nodeIp := range nodeIps {
		result[i].ServerIpv4 = nodeIp
		result[i].ServerId = nodeInfo[nodeIp].ServerId
	}

	for i, msg := range systemStatsResp.GetMessages() {
		stat := msg.CpuTotal
		idle := stat.Idle + stat.Iowait
		nonIdle := stat.User + stat.Nice + stat.System + stat.Irq + stat.Steal + stat.SoftIrq
		total := idle + nonIdle
		cpuUtil := 0.0
		if total > 0 {
			cpuUtil = (total - idle) / total
		}
		result[i].CpuUtilization = cpuUtil
	}

	for i, msg := range memoryResp.GetMessages() {
		memInfo := msg.GetMeminfo()
		memTotal := memInfo.GetMemtotal()
		memUsed := memInfo.GetMemtotal() - memInfo.GetMemfree() - memInfo.GetCached() - memInfo.GetBuffers()

		memUtil := 0.0
		if memTotal > 0 {
			memUtil = float64(memUsed) / float64(memTotal)
		}
		result[i].MemoryUtilization = memUtil
	}
	return result, nil
}

// ServerStatsStream streams ServerStats at the specified interval
func (p *TalosClusterCellProvider) ServerStatsStream(ctx context.Context, cellId string, interval time.Duration) <-chan ServerStatsResult {
	resultChan := make(chan ServerStatsResult)

	go func() {
		defer close(resultChan)

		// Perform one-time setup
		setup, err := p.setupClients(ctx, cellId)
		if err != nil {
			resultChan <- ServerStatsResult{Error: err}
			return
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats, err := p.getServerStatsWithClients(ctx, setup.k8sClient, setup.talosClient, setup.nodeIps)
				resultChan <- ServerStatsResult{Stats: stats, Error: err}
			}
		}
	}()

	return resultChan
}

func (p *TalosClusterCellProvider) ServerStats(ctx context.Context, cellId string) ([]ServerStats, error) {
	setup, err := p.setupClients(ctx, cellId)
	if err != nil {
		return nil, err
	}

	return p.getServerStatsWithClients(ctx, setup.k8sClient, setup.talosClient, setup.nodeIps)
}

func encryptYaml(data string, identity string) (string, error) {
	id, err := age.ParseX25519Identity(identity)
	if err != nil {
		return "", fmt.Errorf("failed to generate identity: %w", err)
	}
	store := sopsyaml.Store{}
	branches, err := store.LoadPlainFile([]byte(data))
	if err != nil {
		return "", fmt.Errorf("failed to load plain file: %w", err)
	}
	masterKey, err := keysource.MasterKeyFromRecipient(id.Recipient().String())
	if err != nil {
		return "", fmt.Errorf("failed to get master key from recipient: %w", err)
	}
	tree := sops.Tree{
		Branches: branches,
		Metadata: sops.Metadata{
			KeyGroups: []sops.KeyGroup{
				[]keys.MasterKey{masterKey},
			},
			UnencryptedSuffix: "_unencrypted",
		},
	}

	dataKey, errs := tree.GenerateDataKeyWithKeyServices(
		[]keyservice.KeyServiceClient{keyservice.NewLocalClient()},
	)
	if errs != nil {
		return "", fmt.Errorf("failed to generate data key: %v", errs)
	}
	if err := common.EncryptTree(common.EncryptTreeOpts{
		DataKey: dataKey,
		Tree:    &tree,
		Cipher:  aes.NewCipher(),
	}); err != nil {
		return "", fmt.Errorf("failed to encrypt tree: %w", err)
	}
	result, err := store.EmitEncryptedFile(tree)
	if err != nil {
		return "", fmt.Errorf("failed to emit encrypted file: %w", err)
	}
	return string(result), nil
}

func archiveRepository(sourceDir, destZip string) error {
	files, err := archiver.FilesFromDisk(nil, map[string]string{
		fmt.Sprintf("%s/", sourceDir): "",
	})
	if err != nil {
		return fmt.Errorf("error gathering files: %v", err)
	}

	out, err := os.Create(destZip)
	if err != nil {
		return fmt.Errorf("error creating zip file: %v", err)
	}
	defer out.Close()

	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}
	err = format.Archive(context.Background(), out, files)
	if err != nil {
		return fmt.Errorf("error archiving files: %v", err)
	}
	return nil
}

func (p *TalosClusterCellProvider) AdvanceDeployment(ctx context.Context, cellId string, deployment *store.Deployment) (*AdvanceDeploymentResult, error) {
	switch deployment.Status {
	case store.DeploymentStatusPending:
		return p.handlePendingDeployment(ctx, cellId, deployment)
	case store.DeploymentStatusDeploying:
		return p.handleDeployingDeployment(ctx, cellId, deployment)
	}
	return nil, nil
}

func (p *TalosClusterCellProvider) DestroyDeployments(ctx context.Context, cellId string, deployments []store.Deployment) error {
	cell, err := p.cellStore.Get(cellId)
	if err != nil {
		return fmt.Errorf("error getting cell: %v", err)
	}
	if cell.TalosCellData == nil {
		return fmt.Errorf("cell %s has no config", cellId)
	}

	clientset, err := initializeK8sClient(cell.TalosCellData.Kubecfg, p.tracerProvider)
	if err != nil {
		return fmt.Errorf("error initializing k8s client: %v", err)
	}

	for _, deployment := range deployments {
		k8sDeployment, err := clientset.AppsV1().Deployments(deployment.Env.Name).Get(ctx, deployment.App.Name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				continue // job's done
			}
			return fmt.Errorf("error getting deployment: %v", err)
		}
		if k8sDeployment.Annotations["onmetal.dev/app-id"] != deployment.App.Id {
			return fmt.Errorf("deployment app id mismatch")
		}
		if k8sDeployment.Annotations["onmetal.dev/team-id"] != deployment.TeamId {
			return fmt.Errorf("deployment team id mismatch")
		}
		if k8sDeployment.Annotations["onmetal.dev/deployment-id"] != fmt.Sprintf("%d", deployment.Id) {
			continue // deployment in k8s is more recent, so we don't need to delete it
		}
		if err := clientset.AppsV1().Deployments(deployment.Env.Name).Delete(ctx, deployment.App.Name, metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf("error deleting deployment: %v", err)
		}
	}
	return nil
}

func (p *TalosClusterCellProvider) handlePendingDeployment(ctx context.Context, cellId string, deployment *store.Deployment) (*AdvanceDeploymentResult, error) {
	// TODO: change error returns into AdvanceDeploymentResult with a failed status and a reason
	cell, err := p.cellStore.Get(cellId)
	if err != nil {
		return nil, fmt.Errorf("error getting cell: %v", err)
	}
	if cell.TalosCellData == nil {
		return nil, fmt.Errorf("cell %s has no config", cellId)
	}

	clientset, err := initializeK8sClient(cell.TalosCellData.Kubecfg, p.tracerProvider)
	if err != nil {
		return nil, fmt.Errorf("error initializing k8s client: %v", err)
	}

	if err := ensureNamespaceExists(ctx, clientset, deployment.Env.Name); err != nil {
		return nil, fmt.Errorf("error ensuring namespace exists: %v", err)
	}

	limits, requests, err := getResourceLimits(deployment.AppSettings.Resources.Data())
	if err != nil {
		return nil, fmt.Errorf("error getting resource limits: %v", err)
	}

	ports, err := getContainerPorts(deployment.AppSettings)
	if err != nil {
		return nil, fmt.Errorf("error getting container ports: %v", err)
	}

	k8sDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployment.App.Name,
			Labels: map[string]string{
				"app": deployment.App.Name,
			},
			Annotations: map[string]string{
				"kubernetes.io/change-cause": fmt.Sprintf("deploy %s id %d", deployment.App.Name, deployment.Id),
				"onmetal.dev/app-id":         deployment.App.Id,
				"onmetal.dev/team-id":        deployment.TeamId,
				"onmetal.dev/deployment-id":  fmt.Sprintf("%d", deployment.Id),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(deployment.Replicas)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deployment.App.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deployment.App.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Resources: corev1.ResourceRequirements{
								Limits:   limits,
								Requests: requests,
							},
							Name:  deployment.App.Name,
							Image: deployment.AppSettings.Artifact.Data().Image.Name,
							Ports: ports,
							Env:   convertEnvVars(deployment.AppEnvVars.EnvVars.Data()),
						},
					},
				},
			},
		},
	}

	// Check if the deployment already exists
	_, err = clientset.AppsV1().Deployments(deployment.Env.Name).Get(ctx, deployment.App.Name, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("error checking existing deployment: %v", err)
		}
		// Deployment doesn't exist, create it
		_, err = clientset.AppsV1().Deployments(deployment.Env.Name).Create(ctx, k8sDeployment, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("error creating deployment: %v", err)
		}
	} else {
		// Deployment exists, update it
		_, err = clientset.AppsV1().Deployments(deployment.Env.Name).Update(ctx, k8sDeployment, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("error updating deployment: %v", err)
		}
	}

	return &AdvanceDeploymentResult{
		Status: store.DeploymentStatusDeploying,
	}, nil
}

func (p *TalosClusterCellProvider) handleDeployingDeployment(ctx context.Context, cellId string, deployment *store.Deployment) (*AdvanceDeploymentResult, error) {
	cell, err := p.cellStore.Get(cellId)
	if err != nil {
		return nil, fmt.Errorf("error getting cell: %v", err)
	}
	if cell.TalosCellData == nil {
		return nil, fmt.Errorf("cell %s has no config", cellId)
	}

	clientset, err := initializeK8sClient(cell.TalosCellData.Kubecfg, p.tracerProvider)
	if err != nil {
		return nil, fmt.Errorf("error initializing k8s client: %v", err)
	}

	// Get the deployment
	k8sDeployment, err := clientset.AppsV1().Deployments(deployment.Env.Name).Get(ctx, deployment.App.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting deployment: %v", err)
	}

	// double check annotations match
	if k8sDeployment.Annotations["onmetal.dev/app-id"] != deployment.App.Id {
		return nil, fmt.Errorf("deployment app id mismatch")
	}
	if k8sDeployment.Annotations["onmetal.dev/team-id"] != deployment.TeamId {
		return nil, fmt.Errorf("deployment team id mismatch")
	}
	if k8sDeployment.Annotations["onmetal.dev/deployment-id"] != fmt.Sprintf("%d", deployment.Id) {
		return nil, fmt.Errorf("deployment deployment id mismatch")
	}

	// Check if the deployment is ready
	if k8sDeployment.Status.ReadyReplicas == k8sDeployment.Status.Replicas {
		return &AdvanceDeploymentResult{
			Status: store.DeploymentStatusRunning,
		}, nil
	}

	// Check for any errors in the deployment
	for _, condition := range k8sDeployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentReplicaFailure && condition.Status == corev1.ConditionTrue {
			return &AdvanceDeploymentResult{
				Status:       store.DeploymentStatusFailed,
				StatusReason: condition.Message,
			}, nil
		}
	}

	// If not ready and no errors, it's still deploying
	return &AdvanceDeploymentResult{
		Status: store.DeploymentStatusDeploying,
	}, nil
}

// initializeK8sClient initializes the Kubernetes client using the provided kubeconfig string.
func initializeK8sClient(kubeconfig string, tp *trace.TracerProvider) (*kubernetes.Clientset, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create REST config from kubeconfig: %w", err)
	}
	config.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		return otelhttp.NewTransport(rt, otelhttp.WithTracerProvider(tp))
	})
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	return clientset, nil
}

// NodeInfo represents the information of a node including its external IP and labels.
type NodeInfo struct {
	Ipv4 string
	NodeLabelInfo
}

// getNodeIpv4ToLabels retrieves node labels and external IPs from the Kubernetes cluster.
func getNodeIpv4ToLabels(ctx context.Context, clientset *kubernetes.Clientset) (map[string]NodeInfo, error) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	nodeInfo := make(map[string]NodeInfo)
	for _, node := range nodes.Items {
		labels := node.Labels
		var ipv4 string
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeExternalIP || addr.Type == corev1.NodeInternalIP {
				ip := addr.Address
				if !strings.HasPrefix(ip, "192.") && !strings.HasPrefix(ip, "172.") && !strings.HasPrefix(ip, "10.") {
					ipv4 = ip
					break
				}
			}
		}
		parsedLabels, err := parseNodeLabels(labels)
		if err != nil {
			return nil, err
		}
		nodeInfo[ipv4] = NodeInfo{
			Ipv4:          ipv4,
			NodeLabelInfo: parsedLabels,
		}
	}
	return nodeInfo, nil
}

type NodeLabelInfo struct {
	ServerId     string
	CellName     string
	ProviderSlug string
}

func generateNodeLabels(info NodeLabelInfo) map[string]string {
	return map[string]string{
		"onmetal.dev/server":   info.ServerId,
		"onmetal.dev/cell":     info.CellName,
		"onmetal.dev/provider": info.ProviderSlug,
	}
}

func parseNodeLabels(labels map[string]string) (NodeLabelInfo, error) {
	info := NodeLabelInfo{
		ServerId:     labels["onmetal.dev/server"],
		CellName:     labels["onmetal.dev/cell"],
		ProviderSlug: labels["onmetal.dev/provider"],
	}

	if info.ServerId == "" || info.CellName == "" || info.ProviderSlug == "" {
		return NodeLabelInfo{}, fmt.Errorf("missing required node label fields")
	}

	return info, nil
}

func ensureNamespaceExists(ctx context.Context, clientset *kubernetes.Clientset, namespace string) error {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Namespace doesn't exist, create it
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			_, err := clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create namespace: %v", err)
			}
			fmt.Printf("Namespace %s created\n", namespace)
		} else {
			return fmt.Errorf("error checking namespace: %v", err)
		}
	}
	return nil
}

func getResourceLimits(resources store.Resources) (corev1.ResourceList, corev1.ResourceList, error) {
	limits := corev1.ResourceList{}
	requests := corev1.ResourceList{}

	cpuLimit, err := resource.ParseQuantity(fmt.Sprintf("%f", resources.Limits.CpuCores))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CPU limit: %v", err)
	}
	memLimit, err := resource.ParseQuantity(fmt.Sprintf("%dMi", resources.Limits.MemoryMiB))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse memory limit: %v", err)
	}

	cpuRequest, err := resource.ParseQuantity(fmt.Sprintf("%f", resources.Requests.CpuCores))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CPU request: %v", err)
	}
	memRequest, err := resource.ParseQuantity(fmt.Sprintf("%dMi", resources.Requests.MemoryMiB))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse memory request: %v", err)
	}

	limits[corev1.ResourceCPU] = cpuLimit
	limits[corev1.ResourceMemory] = memLimit
	requests[corev1.ResourceCPU] = cpuRequest
	requests[corev1.ResourceMemory] = memRequest

	return limits, requests, nil
}

// getContainerPorts converts AppSettings ports to Kubernetes container ports
func getContainerPorts(appSettings store.AppSettings) ([]corev1.ContainerPort, error) {
	ports := appSettings.Ports.Data()
	containerPorts := make([]corev1.ContainerPort, len(ports))
	for i, port := range ports {
		proto, err := getContainerPortProto(port)
		if err != nil {
			return nil, fmt.Errorf("failed to get container port protocol: %w", err)
		}
		containerPorts[i] = corev1.ContainerPort{
			Name:          port.Name,
			ContainerPort: int32(port.Port),
			Protocol:      proto,
		}
	}
	return containerPorts, nil
}

func getContainerPortProto(port store.Port) (corev1.Protocol, error) {
	switch port.Proto {
	case "http":
		return corev1.ProtocolTCP, nil
	}
	return "", fmt.Errorf("invalid port protocol: %s", port.Proto)
}

// convertEnvVars converts []store.EnvVar to []corev1.EnvVar
func convertEnvVars(storeEnvVars []store.EnvVar) []corev1.EnvVar {
	k8sEnvVars := make([]corev1.EnvVar, len(storeEnvVars))
	for i, env := range storeEnvVars {
		k8sEnvVars[i] = corev1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		}
	}
	return k8sEnvVars
}

// func unarchiveRepository(sourceZip, destDir string) error {
//  in, err := os.Open(sourceZip)
//  if err != nil {
//      return fmt.Errorf("error opening zip file: %v", err)
//  }
//  defer in.Close()
//  format := archiver.CompressedArchive{
//      Compression: archiver.Gz{},
//      Archival:    archiver.Tar{},
//  }
//  err = format.Extract(context.Background(), in, nil, func(ctx context.Context, f archiver.File) error {
//      filePath := filepath.Join(destDir, f.NameInArchive)
//      if f.FileInfo.IsDir() {
//          return os.MkdirAll(filePath, os.ModePerm)
//      }
//      destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo.Mode())
//      if err != nil {
//          return err
//      }
//      defer destFile.Close()
//      srcFile, err := f.Open()
//      if err != nil {
//          return err
//      }
//      defer srcFile.Close()
//      _, err = io.Copy(destFile, srcFile)
//      return err
//  })
//  if err != nil {
//      return fmt.Errorf("error extracting files: %v", err)
//  }
//  return nil
// }
