// package cellprovider contains logic to create and manage groups of servers
package cellprovider

import (
	"context"
	"io"
	"time"

	"github.com/onmetal-dev/metal/lib/store"
)

type CreateCellOptions struct {
	Name              string       `valid:"required, matches(^[a-z-]+$)"`
	TeamId            string       `valid:"required"`
	TeamName          string       `valid:"required"`
	TeamAgePrivateKey string       `valid:"required, matches(^AGE-SECRET-KEY.*$)"`
	DnsZoneId         string       `valid:"required"`
	FirstServer       store.Server `valid:"required"`
}

type ServerStats struct {
	ServerId          string
	ServerIpv4        string
	CpuUtilization    float64
	MemoryUtilization float64
}

type AdvanceDeploymentResult struct {
	Status       store.DeploymentStatus
	StatusReason string
}

type ServerStatsResult struct {
	Stats []ServerStats
	Error error
}

type LogEntry struct {
	Timestamp time.Time
	Message   string
}

type DeploymentLogsResult struct {
	Annotations map[string]string
	Logs        []LogEntry
	Error       error
}

// DeploymentLogsOptions defines options for fetching deployment logs
type DeploymentLogsOptions struct {
	Since *time.Duration
}

// DeploymentLogsOption is a function that modifies DeploymentLogsOptions
type DeploymentLogsOption func(*DeploymentLogsOptions)

// WithSince sets the Since option for deployment logs
func WithSince(duration time.Duration) DeploymentLogsOption {
	return func(opts *DeploymentLogsOptions) {
		opts.Since = &duration
	}
}

// processDeploymentLogsOptions applies the given options to DeploymentLogsOptions
func processDeploymentLogsOptions(opts ...DeploymentLogsOption) DeploymentLogsOptions {
	options := DeploymentLogsOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

type BuildImageOptions struct {
	// CellId is the id of the cell to build the image on.
	CellId string `validate:"required"`
	// BuildDir is the directory containing the build context. This is where docker commands will be run.
	BuildDir string `validate:"required"`
	// AppName is the name of the application to build
	AppName string `validate:"required"`
	// BuildId is the id of the build
	BuildId string `validate:"required"`
	// Stderr is the writer to write stderr output to (the build process will exec some docker commands).
	Stderr io.Writer
	// Stdout is the writer to write stdout output to (the build process will exec some docker commands).
	Stdout io.Writer
}

type CellProvider interface {
	CreateCell(ctx context.Context, opts CreateCellOptions) (*store.Cell, error)
	Janitor(ctx context.Context, cellId string) error
	ServerStats(ctx context.Context, cellId string) ([]ServerStats, error)
	ServerStatsStream(ctx context.Context, cellId string, interval time.Duration) <-chan ServerStatsResult
	AdvanceDeployment(ctx context.Context, cellId string, deployment *store.Deployment) (*AdvanceDeploymentResult, error)
	DestroyDeployments(ctx context.Context, cellId string, deployments []store.Deployment) error
	DeploymentLogs(ctx context.Context, cellId string, deployment *store.Deployment, opts ...DeploymentLogsOption) ([]LogEntry, error)
	DeploymentLogsStream(ctx context.Context, cellId string, deployment *store.Deployment, opts ...DeploymentLogsOption) <-chan DeploymentLogsResult
	BuildImage(ctx context.Context, opts BuildImageOptions) (*store.ImageArtifact, error)
}
