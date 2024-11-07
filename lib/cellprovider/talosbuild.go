package cellprovider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/onmetal-dev/metal/lib/validate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *TalosClusterCellProvider) BuildImage(ctx context.Context, opts BuildImageOptions) (*store.ImageArtifact, error) {
	logger := logger.FromContext(ctx).With("cellId", opts.CellId)
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	if opts.Stderr == nil || opts.Stdout == nil {
		return nil, fmt.Errorf("stderr and stdout are required")
	}

	cell, err := c.cellStore.Get(opts.CellId)
	if err != nil {
		return nil, err
	}

	// Create a temporary file for the kubeconfig
	kubeconfigFile, err := os.CreateTemp("", "kubeconfig-*")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary kubeconfig file: %v", err)
	}
	defer os.Remove(kubeconfigFile.Name())
	if _, err := kubeconfigFile.WriteString(cell.TalosCellData.Kubecfg); err != nil {
		return nil, fmt.Errorf("error writing kubeconfig to temporary file: %v", err)
	}
	if err := kubeconfigFile.Close(); err != nil {
		return nil, fmt.Errorf("error closing temporary kubeconfig file: %v", err)
	}

	// log in to the cell's registry
	setup, err := c.setupClients(ctx, opts.CellId)
	if err != nil {
		return nil, fmt.Errorf("failed to setup clients: %w", err)
	}
	secretName := "registry-auth-plaintext"
	namespace := "registry"
	secret, err := setup.k8sClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	username := string(secret.Data["username"])
	password := string(secret.Data["password"])
	registryURL := cellRegistryHostname(opts.CellId)
	cmd := exec.CommandContext(ctx, "docker", "login", "-u", username, "-p", password, registryURL)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to login to docker registry: %w\n%s", err, string(output))
	}

	// find/create the buildx builder
	cmd = exec.CommandContext(ctx, "docker", "buildx", "ls", "--format", "json")
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kubeconfigFile.Name()))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running docker buildx ls: %v", err)
	}
	var builders []BuildxBuilder
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		var builder BuildxBuilder
		if err := json.Unmarshal(scanner.Bytes(), &builder); err != nil {
			return nil, fmt.Errorf("error parsing builder JSON: %v\n%s", err, scanner.Text())
		}
		builders = append(builders, builder)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning builder output: %v", err)
	}
	found := false
	for _, builder := range builders {
		if builder.Name == opts.CellId {
			logger.Info("builder found")
			found = true
			break
		}
	}
	if !found {
		// run docker buildx create --bootstrap --driver kubernetes --name {cellId} --platform=linux/amd64 --node=builder-amd64 --driver-opt=namespace=buildkit,nodeselector="kubernetes.io/arch=amd64"
		logger.Info("creating builder")
		cmd := exec.CommandContext(ctx, "docker", "buildx", "create", "--bootstrap", "--driver", "kubernetes", "--name", opts.CellId, "--platform=linux/amd64", "--node=builder-amd64", "--driver-opt=namespace=buildkit,nodeselector=kubernetes.io/arch=amd64")
		cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kubeconfigFile.Name()))
		cmd.Stderr = opts.Stderr
		cmd.Stdout = opts.Stdout
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("error running docker buildx create: %v\n%s", err, string(output))
		}
	}

	// build the image using buildx
	logger.Info("building image")
	imageRegistry := cellRegistryHostname(opts.CellId)
	imageRepository := opts.AppName
	imageTag := opts.BuildId
	cmd = exec.CommandContext(ctx, "docker", "buildx", "build",
		"-f", "./env/prod/app/Dockerfile", // TODO: make this configurable
		"-t", fmt.Sprintf("%s/%s:%s", imageRegistry, imageRepository, imageTag),
		"--load",
		"--push",
		"--progress", "plain",
		"--builder", opts.CellId,
		".")
	cmd.Dir = opts.BuildDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kubeconfigFile.Name()))
	cmd.Stderr = opts.Stderr
	cmd.Stdout = opts.Stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error running docker buildx build: %w", err)
	}
	return &store.ImageArtifact{
		Registry:   imageRegistry,
		Repository: imageRepository,
		Tag:        imageTag,
	}, nil
}

// BuildxBuilder is the output of docker buildx ls --format json. If the Builder type in github.com/docker/buildx/builder adds an unmarshaljson method, we can remove this.
type BuildxBuilder struct {
	Name         string
	Driver       string
	LastActivity time.Time `json:",omitempty"`
	Dynamic      bool
	Nodes        []BuildxNode
}

type BuildxNode struct {
	Name      string
	Endpoint  string
	Status    string
	Version   string
	Platforms []string
}
