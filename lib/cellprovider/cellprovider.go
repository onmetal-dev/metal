// package cellprovider contains logic to create and manage groups of servers
package cellprovider

import (
	"context"
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

type CellProvider interface {
	CreateCell(ctx context.Context, opts CreateCellOptions) (*store.Cell, error)
	ServerStats(ctx context.Context, cellId string) ([]ServerStats, error)
	ServerStatsStream(ctx context.Context, cellId string, interval time.Duration) <-chan ServerStatsResult
	AdvanceDeployment(ctx context.Context, cellId string, deployment *store.Deployment) (*AdvanceDeploymentResult, error)
	DestroyDeployments(ctx context.Context, cellId string, deployments []store.Deployment) error
}
