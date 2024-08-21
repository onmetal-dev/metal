package cellprovider

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
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
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/onmetal-dev/metal/lib/dnsprovider"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/siderolabs/talos/cmd/talosctl/pkg/talos/helpers"
	"github.com/siderolabs/talos/pkg/machinery/api/machine"
	"github.com/siderolabs/talos/pkg/machinery/api/storage"
	"github.com/siderolabs/talos/pkg/machinery/client"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
	"github.com/siderolabs/talos/pkg/machinery/config"
	"github.com/siderolabs/talos/pkg/machinery/config/generate/secrets"
)

// TalosClusterCellProvider creates a talos k8s cluster using cloudflare for DNS
type TalosClusterCellProvider struct {
	dnsProvider dnsprovider.DNSProvider
	cellStore   store.CellStore
	serverStore store.ServerStore
	tmpDirRoot  string
	logger      *slog.Logger
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
	c, err := client.New(ctx, client.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}), client.WithEndpoints(*opts.FirstServer.PublicIpv4))
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
					NodeLabels: map[string]string{
						"onmetal.dev/server":   opts.FirstServer.Id,
						"onmetal.dev/cell":     opts.Name,
						"onmetal.dev/provider": opts.FirstServer.ProviderSlug,
					},
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

func (p *TalosClusterCellProvider) ServerStats(ctx context.Context, cellId string) ([]ServerStats, error) {
	cell, err := p.cellStore.Get(cellId)
	if err != nil {
		return nil, fmt.Errorf("error getting cell: %v", err)
	}

	clientConfig, err := clientconfig.FromString(cell.TalosCellData.Talosconfig)
	if err != nil {
		return nil, fmt.Errorf("error parsing talosconfig: %v", err)
	}

	c, err := client.New(ctx,
		client.WithConfig(clientConfig),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create talos client from config: %w", err)
	}

	resp, err := c.MachineClient.SystemStat(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("error getting system stats: %v", err)
	}

	result := make([]ServerStats, len(resp.GetMessages()))
	for i, msg := range resp.GetMessages() {
		// heavily borrowed from https://github.com/siderolabs/talos/blob/36f83eea9f6baba358c1d98223a330b2cb26e988/internal/pkg/dashboard/apidata/node.go#L52
		stat := msg.CpuTotal
		idle := stat.Idle + stat.Iowait
		nonIdle := stat.User + stat.Nice + stat.System + stat.Irq + stat.Steal + stat.SoftIrq
		total := idle + nonIdle
		cpuUtil := 0.0
		if total > 0 {
			cpuUtil = (total - idle) / total
		}
		// TODO: for some reason this is blank
		//hostname := msg.GetMetadata().GetHostname()
		result[i] = ServerStats{
			CpuUtilization: cpuUtil,
		}
	}

	respMem, err := c.MachineClient.Memory(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("error getting memory stats: %v", err)
	}
	for i, msg := range respMem.GetMessages() {
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

// func unarchiveRepository(sourceZip, destDir string) error {
// 	in, err := os.Open(sourceZip)
// 	if err != nil {
// 		return fmt.Errorf("error opening zip file: %v", err)
// 	}
// 	defer in.Close()

// 	format := archiver.CompressedArchive{
// 		Compression: archiver.Gz{},
// 		Archival:    archiver.Tar{},
// 	}
// 	err = format.Extract(context.Background(), in, nil, func(ctx context.Context, f archiver.File) error {
// 		filePath := filepath.Join(destDir, f.NameInArchive)
// 		if f.FileInfo.IsDir() {
// 			return os.MkdirAll(filePath, os.ModePerm)
// 		}

// 		destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo.Mode())
// 		if err != nil {
// 			return err
// 		}
// 		defer destFile.Close()

// 		srcFile, err := f.Open()
// 		if err != nil {
// 			return err
// 		}
// 		defer srcFile.Close()

// 		_, err = io.Copy(destFile, srcFile)
// 		return err
// 	})
// 	if err != nil {
// 		return fmt.Errorf("error extracting files: %v", err)
// 	}

// 	return nil
// }
