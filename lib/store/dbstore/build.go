package dbstore

import (
	"context"
	"fmt"

	"github.com/onmetal-dev/metal/lib/store"
	"go.jetify.com/typeid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BuildStore struct {
	db *gorm.DB
}

var _ store.BuildStore = &BuildStore{}

func NewBuildStore(db *gorm.DB) *BuildStore {
	return &BuildStore{db: db}
}

func (s *BuildStore) Init(ctx context.Context, opts store.InitBuildOptions) (store.Build, error) {
	tid, _ := typeid.WithPrefix("build")
	build := store.Build{
		Common: store.Common{
			Id: tid.String(),
		},
		TeamId:    opts.TeamId,
		CreatorId: opts.CreatorId,
		AppId:     opts.AppId,
		Status:    store.BuildStatusPending,
	}

	if err := s.db.WithContext(ctx).Create(&build).Error; err != nil {
		return store.Build{}, fmt.Errorf("failed to create build: %w", err)
	}

	return build, nil
}

func (s *BuildStore) Get(ctx context.Context, id string) (store.Build, error) {
	var build store.Build
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&build).Error; err != nil {
		return store.Build{}, fmt.Errorf("failed to get build: %w", err)
	}
	return build, nil
}

func (s *BuildStore) UpdateStatus(ctx context.Context, id string, status store.BuildStatus, statusReason string) error {
	if err := s.db.WithContext(ctx).Model(&store.Build{}).Where("id = ?", id).Update("status", status).Update("status_reason", statusReason).Error; err != nil {
		return fmt.Errorf("failed to update build status: %w", err)
	}
	return nil
}

func (s *BuildStore) UpdateLogs(ctx context.Context, id string, logs store.BuildLogs) error {
	if err := s.db.WithContext(ctx).Model(&store.Build{}).Where("id = ?", id).Update("logs", datatypes.NewJSONType(logs)).Error; err != nil {
		return fmt.Errorf("failed to update build logs: %w", err)
	}
	return nil
}

func (s *BuildStore) UpdateArtifacts(ctx context.Context, id string, artifacts []store.Artifact) error {
	if err := s.db.WithContext(ctx).Model(&store.Build{}).Where("id = ?", id).Update("artifacts", datatypes.NewJSONType(artifacts)).Error; err != nil {
		return fmt.Errorf("failed to update build artifacts: %w", err)
	}
	return nil
}
