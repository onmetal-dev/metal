package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	thconfig "github.com/budimanjojo/talhelper/v3/pkg/config"
	"github.com/budimanjojo/talhelper/v3/pkg/generate"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-yaml/yaml"
	"github.com/kelseyhightower/envconfig"
	"github.com/mholt/archiver/v4"
	"github.com/onmetal-dev/metal/lib/store"
	database "github.com/onmetal-dev/metal/lib/store/db"
	"github.com/onmetal-dev/metal/lib/store/dbstore"
	"github.com/siderolabs/talos/cmd/talosctl/pkg/talos/helpers"
	"github.com/siderolabs/talos/pkg/machinery/api/machine"
	"github.com/siderolabs/talos/pkg/machinery/client"
)

type Config struct {
	DatabaseHost     string `envconfig:"DATABASE_HOST" default:"localhost" required:"true"`
	DatabasePort     int    `envconfig:"DATABASE_PORT" default:"5432" required:"true"`
	DatabaseUser     string `envconfig:"DATABASE_USER" default:"postgres" required:"true"`
	DatabasePassword string `envconfig:"DATABASE_PASSWORD" default:"postgres" required:"true"`
	DatabaseName     string `envconfig:"DATABASE_NAME" default:"metal" required:"true"`
	DatabaseSslMode  string `envconfig:"DATABASE_SSL_MODE" default:"disable" required:"true"`
	TmpDirRoot       string `envconfig:"TMP_DIR_ROOT" required:"true"`
	ServerId         string `envconfig:"SERVER_ID" required:"true"`
	CellId           string `envconfig:"CELL_ID" required:"true"`
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

// WithEnv should eventually be factored out into a global lib w/ a global mutex for env vars
func WithEnv(m map[string]string, f func() error) error {
	for k, v := range m {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range m {
			os.Unsetenv(k)
		}
	}()
	return f()
}

func run(ctx context.Context) error {
	c := MustLoadConfig()

	db := database.MustOpen(c.DatabaseHost, c.DatabaseUser, c.DatabasePassword, c.DatabaseName, c.DatabasePort, c.DatabaseSslMode, nil)
	cellStore := dbstore.NewCellStore(
		dbstore.NewCellStoreParams{
			DB: db,
		},
	)
	teamStore := dbstore.NewTeamStore(
		dbstore.NewTeamStoreParams{
			DB: db,
		},
	)
	cell, err := cellStore.Get(c.CellId)
	if err != nil {
		return fmt.Errorf("error getting cell: %v", err)
	}

	team, err := teamStore.GetTeam(ctx, cell.TeamId)
	if err != nil {
		return fmt.Errorf("error getting team: %v", err)
	}
	teamSopsKey := team.AgePrivateKey

	// Unzip the repository to a new temporary directory
	unzipDir, err := os.MkdirTemp(c.TmpDirRoot, "unarchived")
	if err != nil {
		return fmt.Errorf("error creating unzip directory: %v", err)
	}
	//defer os.RemoveAll(unzipDir)

	if err := unarchiveRepository(cell.TalosCellData.Config, unzipDir); err != nil {
		return fmt.Errorf("error unarchiving repository: %v", err)
	}
	secretFilePath := filepath.Join(unzipDir, "talsecret.sops.yaml")

	configFilePath := filepath.Join(unzipDir, "talconfig.yaml")
	thConfig, err := thconfig.LoadAndValidateFromFile(configFilePath, nil, false)
	if err != nil {
		return fmt.Errorf("error loading talconfig: %v", err)
	}

	newPatches := []string{}
	kubeletPatch := `- op: add
  path: /machine/kubelet/extraArgs
  value:
    rotate-server-certificates: "true"`
	for _, p := range thConfig.Patches {
		if p == kubeletPatch {
			continue
		}
		newPatches = append(newPatches, p)
	}
	thConfig.Patches = append(newPatches, kubeletPatch)
	thConfig.Nodes[0].NodeConfigs.NodeLabels["onmetal.dev/server"] = c.ServerId

	// write the config back to disk
	configData, err := yaml.Marshal(thConfig)
	if err != nil {
		return fmt.Errorf("error marshalling talconfig: %v", err)
	}
	if err := os.WriteFile(configFilePath, configData, 0644); err != nil {
		return fmt.Errorf("error writing talconfig to disk: %v", err)
	}

	if err := WithEnv(map[string]string{
		"SOPS_AGE_KEY": teamSopsKey,
	}, func() error {
		return generate.GenerateConfig(thConfig, false, filepath.Join(unzipDir, "clusterconfig"), secretFilePath, "metal", false)
	}); err != nil {
		return fmt.Errorf("error generating config: %v", err)
	}

	cpConfigFilePath := filepath.Join(unzipDir, "clusterconfig", fmt.Sprintf("%s-cp-1.yaml", thConfig.ClusterName))
	cfgBytes, err := os.ReadFile(cpConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to read configuration from %s: %w", cpConfigFilePath, err)
	} else if len(cfgBytes) < 1 {
		return errors.New("no configuration data read")
	}

	talosClient, err := client.New(ctx,
		client.WithConfigFromFile(filepath.Join(unzipDir, "clusterconfig", "talosconfig")),
		client.WithContextName(thConfig.ClusterName),
	)
	if err != nil {
		return fmt.Errorf("failed to create talos client from config: %w", err)
	}
	resp, err := talosClient.ApplyConfiguration(ctx, &machine.ApplyConfigurationRequest{
		Data: cfgBytes,
	})
	if err != nil {
		return fmt.Errorf("error applying new configuration: %s", err)
	}

	helpers.PrintApplyResults(resp)

	repo, err := git.PlainOpen(unzipDir)
	if err != nil {
		return fmt.Errorf("error opening git repository: %v", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %v", err)
	}
	if err := worktree.AddWithOptions(&git.AddOptions{
		All: true,
	}); err != nil {
		return fmt.Errorf("error adding files to repository: %v", err)
	}
	_, err = worktree.Commit("Update configuration", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Metal",
			Email: "automated@onmetal.dev",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("error committing changes: %v", err)
	}

	// save archive to cell
	archive, err := archiveRepository(unzipDir)
	if err != nil {
		return fmt.Errorf("error archiving repository: %v", err)
	}
	talosConfig, err := os.ReadFile(filepath.Join(unzipDir, "clusterconfig", "talosconfig"))
	if err != nil {
		return fmt.Errorf("error reading talosconfig file: %v", err)
	}
	kubecfg, err := talosClient.Kubeconfig(ctx)
	if err != nil {
		return fmt.Errorf("error getting kubeconfig: %v", err)
	}

	if err := cellStore.UpdateTalosCellData(&store.TalosCellData{
		CellId:      cell.Id,
		Talosconfig: string(talosConfig),
		Config:      archive,
		Kubecfg:     string(kubecfg),
	}); err != nil {
		return fmt.Errorf("error updating cell: %v", err)
	}
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func archiveRepository(sourceDir string) ([]byte, error) {
	files, err := archiver.FilesFromDisk(nil, map[string]string{
		fmt.Sprintf("%s/", sourceDir): "",
	})
	if err != nil {
		return nil, fmt.Errorf("error gathering files: %v", err)
	}

	var buf bytes.Buffer

	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}
	err = format.Archive(context.Background(), &buf, files)
	if err != nil {
		return nil, fmt.Errorf("error archiving files: %v", err)
	}
	return buf.Bytes(), nil
}

func unarchiveRepository(sourceZip []byte, destDir string) error {
	in := bytes.NewReader(sourceZip)

	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}
	err := format.Extract(context.Background(), in, nil, func(ctx context.Context, f archiver.File) error {
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
