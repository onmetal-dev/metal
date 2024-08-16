package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"filippo.io/age"
	"github.com/asaskevich/govalidator"
	thconfig "github.com/budimanjojo/talhelper/v3/pkg/config"
	"github.com/budimanjojo/talhelper/v3/pkg/generate"
	"github.com/cloudflare/cloudflare-go"
	"github.com/davecgh/go-spew/spew"
	"github.com/floshodan/hrobot-go/hrobot"
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

	"github.com/kelseyhightower/envconfig"
	"github.com/onmetal-dev/metal/lib/talosprovider"
	"github.com/siderolabs/talos/cmd/talosctl/pkg/talos/helpers"
	"github.com/siderolabs/talos/pkg/machinery/api/machine"
	"github.com/siderolabs/talos/pkg/machinery/api/storage"
	"github.com/siderolabs/talos/pkg/machinery/client"
	"github.com/siderolabs/talos/pkg/machinery/config"
	"github.com/siderolabs/talos/pkg/machinery/config/generate/secrets"
)

type Config struct {
	TmpDirRoot                    string `envconfig:"TMP_DIR_ROOT" required:"true"`
	HetznerToken                  string `envconfig:"HETZNER_TOKEN" required:"true"`
	SshKeyBase64                  string `envconfig:"SSH_KEY_BASE64" required:"true"`
	SshKeyPassword                string `envconfig:"SSH_KEY_PASSWORD" required:"true"`
	SshKeyFingerprint             string `envconfig:"SSH_KEY_FINGERPRINT" required:"true"`
	CloudflareApiToken            string `envconfig:"CLOUDFLARE_API_TOKEN" required:"true"`
	CloudflareOnmetalDotRunZoneId string `envconfig:"CLOUDFLARE_ONMETAL_DOT_RUN_ZONE_ID" required:"true"`
	TeamName                      string `envconfig:"TEAM_NAME" required:"true"`
	TeamAgeSecretKey              string `envconfig:"TEAM_AGE_SECRET_KEY" required:"true"`
	ServerProviderSlug            string `envconfig:"SERVER_PROVIDER_SLUG" required:"true"`
	ServerProviderId              string `envconfig:"SERVER_PROVIDER_ID" required:"true"`
	ServerId                      string `envconfig:"SERVER_ID" required:"true"`
	ServerIp                      string `envconfig:"SERVER_IP" required:"true"`
}

func loadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func MustLoadConfig() *Config {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

type DNSProvider interface {
	Domain() (string, error)
	FindOrCreateARecord(ctx context.Context, zoneID, recordName, recordContent string) error
}

type CloudflareDNSProvider struct {
	api      *cloudflare.API
	zoneId   string
	zoneName *string
}

var _ DNSProvider = &CloudflareDNSProvider{}

type CloudflareDNSProviderOption func(*CloudflareDNSProvider)

func WithApi(api *cloudflare.API) CloudflareDNSProviderOption {
	return func(p *CloudflareDNSProvider) {
		p.api = api
	}
}

func WithZoneId(zoneId string) CloudflareDNSProviderOption {
	return func(p *CloudflareDNSProvider) {
		p.zoneId = zoneId
	}
}

func NewCloudflareDNSProvider(opts ...CloudflareDNSProviderOption) (*CloudflareDNSProvider, error) {
	provider := &CloudflareDNSProvider{}
	for _, opt := range opts {
		opt(provider)
	}
	var errs []string
	if provider.api == nil {
		errs = append(errs, "must provide a valid Cloudflare API")
	}
	if provider.zoneId == "" {
		errs = append(errs, "must provide a valid zoneId")
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("errors: %v", strings.Join(errs, ", "))
	}
	return provider, nil
}

func (p *CloudflareDNSProvider) Domain() (string, error) {
	if p.zoneName == nil {
		zone, err := p.api.ZoneDetails(context.Background(), p.zoneId)
		if err != nil {
			return "", err
		}
		p.zoneName = &zone.Name
	}
	return *p.zoneName, nil
}

func (p *CloudflareDNSProvider) FindOrCreateARecord(ctx context.Context, zoneID, recordName, recordContent string) error {
	domain, err := p.Domain()
	if err != nil {
		return err
	}
	dnsRecords, _, err := p.api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{
		Type: "A",
		Name: fmt.Sprintf("%s.%s", recordName, domain),
	})
	if err != nil {
		return fmt.Errorf("error listing DNS records: %v", err)
	} else if len(dnsRecords) > 0 {
		if dnsRecords[0].Content != recordContent {
			return fmt.Errorf("existing record content mismatch: %s != %s", dnsRecords[0].Content, recordContent)
		}
		return nil
	}
	if _, err = p.api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
		Type:    "A",
		Name:    recordName,
		Content: recordContent,
	}); err != nil {
		return fmt.Errorf("error creating A record: %v", err)
	}
	return nil
}

// Cell is a group of servers. It aims to enable a cell-based architecture where different tenants, environments, and / or use cases are isolated from each other.
type Cell struct {
	Id        string
	Name      string
	TeamId    string
	ServerIps []string
	ServerIds []string
}

type CellServer struct {
	Id       string `valid:"required, matches(^server_.*$)"`
	Ip       string `valid:"required, ipv4"`
	Provider string `valid:"required"`
}

type CreateCellOptions struct {
	Name            string `valid:"required, matches(^[a-z-]+$)"`
	TeamName        string `valid:"required"`
	TeamIdentity    string `valid:"required, matches(^AGE-SECRET-KEY.*$)"`
	DnsZoneId       string `valid:"required"`
	ReinstallServer *talosprovider.Server
	FirstServer     CellServer `valid:"required"`
}

type CellProvider interface {
	CreateCell(ctx context.Context, opts CreateCellOptions) (*Cell, error)
}

// TalosClusterCellProvider creates a talos k8s cluster using cloudflare for DNS
type TalosClusterCellProvider struct {
	dnsProvider   DNSProvider
	talosProvider talosprovider.TalosProvider
	tmpDirRoot    string
	logger        *slog.Logger
}

var _ CellProvider = &TalosClusterCellProvider{}

type TalosClusterCellProviderOption func(*TalosClusterCellProvider)

func WithDnsProvider(dnsProvider DNSProvider) TalosClusterCellProviderOption {
	return func(p *TalosClusterCellProvider) {
		p.dnsProvider = dnsProvider
	}
}

func WithTalosProvider(talosProvider talosprovider.TalosProvider) TalosClusterCellProviderOption {
	return func(p *TalosClusterCellProvider) {
		p.talosProvider = talosProvider
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
	if provider.talosProvider == nil {
		errs = append(errs, fmt.Errorf("must provide a valid talos provider"))
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
func (p *TalosClusterCellProvider) CreateCell(ctx context.Context, opts CreateCellOptions) (*Cell, error) {
	talosVersion := "1.7.6"
	if _, err := govalidator.ValidateStruct(opts); err != nil {
		return nil, fmt.Errorf("error validating createcell options: %v", err)
	}

	if opts.ReinstallServer != nil {
		if err := p.talosProvider.Install(ctx, *opts.ReinstallServer,
			talosprovider.WithTalosVersion(talosVersion),
			talosprovider.WithArch("amd64")); err != nil {
			return nil, fmt.Errorf("error installing server: %v", err)
		}
	}

	if err := p.dnsProvider.FindOrCreateARecord(ctx, opts.DnsZoneId, opts.FirstServer.Id, opts.FirstServer.Ip); err != nil {
		return nil, err
	}
	domain, err := p.dnsProvider.Domain()
	if err != nil {
		return nil, err
	}

	// figure out install disk
	c, err := client.New(ctx, client.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}), client.WithEndpoints(opts.FirstServer.Ip))
	if err != nil {
		return nil, fmt.Errorf("failed to create talos client: %w", err)
	}
	var disksResp *storage.DisksResponse
	for {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		disksResp, err = c.Disks(ctxWithTimeout)
		if err == nil {
			break
		} else {
			if opts.ReinstallServer != nil {
				// takes time for it to come up
				p.logger.Info("waiting for server to come up after reinstall")
				time.Sleep(10 * time.Second)
				continue
			}
			return nil, fmt.Errorf("failed to get disks: %w", err)
		}
	}
	systemDisk, ok := lo.Find(disksResp.Messages[0].Disks, func(d *storage.Disk) bool {
		return d.SystemDisk
	})
	if !ok {
		return nil, fmt.Errorf("system disk not found")
	}

	// craft the single-node talos config
	thConfig := thconfig.TalhelperConfig{
		ClusterName:              opts.Name,
		TalosVersion:             talosVersion,
		Endpoint:                 fmt.Sprintf("https://%s.%s:6443", opts.FirstServer.Id, domain),
		AllowSchedulingOnMasters: true,
		KubernetesVersion:        "v1.30.3",
		Patches: []string{
			`- op: add
  path: /cluster/discovery/enabled
  value: true
- op: replace
  path: /machine/network/kubespan
  value:
    enabled: true`,
		},
		Nodes: []thconfig.Node{
			{
				Hostname:     "cp-1",
				IPAddress:    opts.FirstServer.Ip,
				ControlPlane: true,
				InstallDisk:  systemDisk.DeviceName,
				NodeConfigs: thconfig.NodeConfigs{
					NodeLabels: map[string]string{
						"onmetal.dev/cell":     opts.Name,
						"onmetal.dev/provider": opts.FirstServer.Provider,
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
	encryptedSecretsBundleStr, err := encryptYaml(secretsBundleStr, opts.TeamIdentity)
	if err != nil {
		return nil, fmt.Errorf("error encrypting secrets bundle: %w", err)
	}
	fmt.Println(encryptedSecretsBundleStr)

	// set up a git repo that will be used to hold cluster config
	tempDir, err := os.MkdirTemp(p.tmpDirRoot, "cluster")
	if err != nil {
		return nil, fmt.Errorf("error creating temp directory: %v", err)
	}
	//	defer os.RemoveAll(tempDir)
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
	os.Setenv("SOPS_AGE_KEY", opts.TeamIdentity)
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
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:50000", opts.FirstServer.Ip), 5*time.Second)
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

	return nil, nil
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

func run(ctx context.Context) error {
	c := MustLoadConfig()
	api, err := cloudflare.NewWithAPIToken(c.CloudflareApiToken)
	if err != nil {
		return fmt.Errorf("error initializing Cloudflare API: %v", err)
	}
	dnsProvider, err := NewCloudflareDNSProvider(WithApi(api), WithZoneId(c.CloudflareOnmetalDotRunZoneId))
	if err != nil {
		return fmt.Errorf("error in NewCloudflareDNSProvider: %v", err)
	}

	hrobotClient := hrobot.NewClient(hrobot.WithToken(c.HetznerToken))
	talosProvider, err := talosprovider.NewHetznerProvider(
		talosprovider.WithClient(hrobotClient),
		talosprovider.WithLogger(slog.Default()),
	)
	if err != nil {
		return fmt.Errorf("error in NewHetznerProvider: %v", err)
	}

	cellProvider, err := NewTalosClusterCellProvider(
		WithDnsProvider(dnsProvider),
		WithTalosProvider(talosProvider),
		WithTmpDirRoot(c.TmpDirRoot),
		WithLogger(slog.Default()),
	)
	if err != nil {
		return fmt.Errorf("error in NewTalosClusterCellProvider: %v", err)
	}

	cell, err := cellProvider.CreateCell(ctx, CreateCellOptions{
		Name:         "default",
		TeamName:     c.TeamName,
		TeamIdentity: c.TeamAgeSecretKey,
		DnsZoneId:    c.CloudflareOnmetalDotRunZoneId,
		ReinstallServer: &talosprovider.Server{
			Id:                    c.ServerProviderId,
			Ip:                    c.ServerIp,
			Username:              "root",
			SshKeyPrivateBase64:   c.SshKeyBase64,
			SshKeyPrivatePassword: c.SshKeyPassword,
			SshKeyFingerprint:     c.SshKeyFingerprint,
		},
		FirstServer: CellServer{
			Id:       c.ServerId,
			Ip:       c.ServerIp,
			Provider: c.ServerProviderSlug,
		},
	})
	if err != nil {
		return fmt.Errorf("error in CreateCell: %v", err)
	}
	spew.Dump(cell)

	os.Exit(0)

	// create a new git repo for the cluster
	tempDir, err := os.MkdirTemp(c.TmpDirRoot, "cluster")
	if err != nil {
		return fmt.Errorf("error creating temp directory: %v", err)
	}
	//	defer os.RemoveAll(tempDir)

	// Initialize a new Git repository
	repo, err := git.PlainInit(tempDir, false)
	if err != nil {
		return fmt.Errorf("error initializing git repository: %v", err)
	}

	// Create a new file in the repository
	filePath := filepath.Join(tempDir, "example.txt")
	err = os.WriteFile(filePath, []byte("Hello, GitOps!"), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	// Add the file to the repository
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %v", err)
	}
	_, err = worktree.Add("example.txt")
	if err != nil {
		return fmt.Errorf("error adding file to repository: %v", err)
	}

	// Commit the changes
	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Metal",
			Email: "automated@onmetal.dev",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("error committing changes: %v", err)
	}

	// Create a new temporary directory for the archive
	archiveTempDir, err := os.MkdirTemp(c.TmpDirRoot, "archive")
	if err != nil {
		return fmt.Errorf("error creating archive temp directory: %v", err)
	}
	//defer os.RemoveAll(archiveTempDir)

	// Archive the repository to a zip file in the new temp directory
	zipFilePath := filepath.Join(archiveTempDir, "repository.zip")
	err = archiveRepository(tempDir, zipFilePath)
	if err != nil {
		return fmt.Errorf("error archiving repository: %v", err)
	}

	// Unzip the repository to a new temporary directory
	unzipDir, err := os.MkdirTemp(c.TmpDirRoot, "unarchived")
	if err != nil {
		return fmt.Errorf("error creating unzip directory: %v", err)
	}
	//defer os.RemoveAll(unzipDir)

	err = unarchiveRepository(zipFilePath, unzipDir)
	if err != nil {
		return fmt.Errorf("error unarchiving repository: %v", err)
	}

	fmt.Println("Repository archived and unarchived successfully.")
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
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

func unarchiveRepository(sourceZip, destDir string) error {
	in, err := os.Open(sourceZip)
	if err != nil {
		return fmt.Errorf("error opening zip file: %v", err)
	}
	defer in.Close()

	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}
	err = format.Extract(context.Background(), in, nil, func(ctx context.Context, f archiver.File) error {
		filePath := filepath.Join(destDir, f.NameInArchive)
		if f.FileInfo.IsDir() {
			return os.MkdirAll(filePath, os.ModePerm)
		}

		destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()

		srcFile, err := f.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
	if err != nil {
		return fmt.Errorf("error extracting files: %v", err)
	}

	return nil
}
