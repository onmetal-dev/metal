package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/davecgh/go-spew/spew"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mholt/archiver/v4"

	"github.com/kelseyhightower/envconfig"
	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/dnsprovider"
	"github.com/onmetal-dev/metal/lib/store"
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

func run(ctx context.Context) error {
	c := MustLoadConfig()
	api, err := cloudflare.NewWithAPIToken(c.CloudflareApiToken)
	if err != nil {
		return fmt.Errorf("error initializing Cloudflare API: %v", err)
	}
	dnsProvider, err := dnsprovider.NewCloudflareDNSProvider(dnsprovider.WithApi(api), dnsprovider.WithZoneId(c.CloudflareOnmetalDotRunZoneId))
	if err != nil {
		return fmt.Errorf("error in NewCloudflareDNSProvider: %v", err)
	}

	// hrobotClient := hrobot.NewClient(hrobot.WithToken(c.HetznerToken))
	// // talosProvider, err := talosprovider.NewHetznerProvider(
	// // 	talosprovider.WithClient(hrobotClient),
	// // 	talosprovider.WithLogger(slog.Default()),
	// // )
	// if err != nil {
	// 	return fmt.Errorf("error in NewHetznerProvider: %v", err)
	// }

	cellProvider, err := cellprovider.NewTalosClusterCellProvider(
		cellprovider.WithDnsProvider(dnsProvider),
		//cellprovider.WithTalosProvider(talosProvider),
		cellprovider.WithTmpDirRoot(c.TmpDirRoot),
		cellprovider.WithLogger(slog.Default()),
	)
	if err != nil {
		return fmt.Errorf("error in NewTalosClusterCellProvider: %v", err)
	}

	cell, err := cellProvider.CreateCell(ctx, cellprovider.CreateCellOptions{
		Name:              "default",
		TeamName:          c.TeamName,
		TeamAgePrivateKey: c.TeamAgeSecretKey,
		DnsZoneId:         c.CloudflareOnmetalDotRunZoneId,
		// ReinstallServer: &store.Server{
		// 	Id:                    c.ServerProviderId,
		// 	Ip:                    c.ServerIp,
		// 	Username:              "root",
		// 	SshKeyPrivateBase64:   c.SshKeyBase64,
		// 	SshKeyPrivatePassword: c.SshKeyPassword,
		// 	SshKeyFingerprint:     c.SshKeyFingerprint,
		// },
		FirstServer: store.Server{
			ProviderId:   &c.ServerId,
			PublicIpv4:   &c.ServerIp,
			ProviderSlug: c.ServerProviderSlug,
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
