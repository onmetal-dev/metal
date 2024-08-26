package dbstore

import (
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/onmetal-dev/metal/lib/validate"
	"github.com/samber/lo"
	"go.jetify.com/typeid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DeploymentStore struct {
	db *gorm.DB
}

var _ store.DeploymentStore = &DeploymentStore{}

type NewDeploymentStoreParams struct {
	DB *gorm.DB
}

func NewDeploymentStore(params NewDeploymentStoreParams) *DeploymentStore {
	return &DeploymentStore{
		db: params.DB,
	}
}

func (s *DeploymentStore) CreateEnv(opts store.CreateEnvOptions) (store.Env, error) {
	tid, _ := typeid.WithPrefix("env")
	if err := validate.Struct(opts); err != nil {
		return store.Env{}, err
	}
	env := store.Env{
		Common: store.Common{Id: tid.String()},
		TeamId: opts.TeamId,
		Name:   opts.Name,
	}
	return env, s.db.Create(&env).Error
}

func (s *DeploymentStore) GetEnv(id string) (store.Env, error) {
	env := store.Env{Common: store.Common{Id: id}}
	return env, s.db.First(&env).Error
}

func (s *DeploymentStore) GetEnvsForTeam(teamId string) ([]store.Env, error) {
	var envs []store.Env
	return envs, s.db.Where(&store.Env{TeamId: teamId}).Find(&envs).Error
}

func (s *DeploymentStore) DeleteEnv(id string) error {
	return s.db.Delete(&store.Env{Common: store.Common{Id: id}}).Error
}

func (s *DeploymentStore) CreateAppEnvVars(opts store.CreateAppEnvVarOptions) (store.AppEnvVars, error) {
	tid, _ := typeid.WithPrefix("appenvvars")
	appEnvVars := store.AppEnvVars{
		Common:  store.Common{Id: tid.String()},
		TeamId:  opts.TeamId,
		EnvId:   opts.EnvId,
		AppId:   opts.AppId,
		EnvVars: datatypes.NewJSONType(opts.EnvVars),
	}
	return appEnvVars, s.db.Create(&appEnvVars).Error
}

func (s *DeploymentStore) GetAppEnvVars(id string) (store.AppEnvVars, error) {
	appEnvVars := store.AppEnvVars{Common: store.Common{Id: id}}
	return appEnvVars, s.db.First(&appEnvVars).Error
}

func (s *DeploymentStore) GetAppEnvVarsForAppEnv(appId string, envId string) ([]store.AppEnvVars, error) {
	var appEnvVars []store.AppEnvVars
	return appEnvVars, s.db.Where(&store.AppEnvVars{AppId: appId, EnvId: envId}).Find(&appEnvVars).Error
}

func (s *DeploymentStore) DeleteAppEnvVars(id string) error {
	return s.db.Delete(&store.AppEnvVars{Common: store.Common{Id: id}}).Error
}

func (s *DeploymentStore) Create(opts store.CreateDeploymentOptions) (store.Deployment, error) {
	deployment := store.Deployment{
		EnvId:         opts.EnvId,
		AppId:         opts.AppId,
		TeamId:        opts.TeamId,
		Type:          opts.Type,
		AppSettingsId: opts.AppSettingsId,
		AppEnvVarsId:  opts.AppEnvVarsId,
		Cells: lo.Map(opts.CellIds, func(cellId string, _ int) store.Cell {
			return store.Cell{Common: store.Common{Id: cellId}}
		}),
	}
	return deployment, s.db.Create(&deployment).Error
}

func (s *DeploymentStore) preloadDeployment(query *gorm.DB) *gorm.DB {
	return query.Preload("Env").Preload("App").Preload("AppSettings").Preload("AppEnvVars").Preload("Cells")
}

func (s *DeploymentStore) Get(appId string, envId string, id uint) (store.Deployment, error) {
	deployment := store.Deployment{Id: id, AppId: appId, EnvId: envId}
	return deployment, s.preloadDeployment(s.db).First(&deployment).Error
}

func (s *DeploymentStore) GetForTeam(teamId string) ([]store.Deployment, error) {
	var deployments []store.Deployment
	return deployments, s.preloadDeployment(s.db).
		Where(&store.Deployment{TeamId: teamId}).
		Find(&deployments).Error
}

func (s *DeploymentStore) GetForApp(appId string) ([]store.Deployment, error) {
	var deployments []store.Deployment
	return deployments, s.preloadDeployment(s.db).
		Where(&store.Deployment{AppId: appId}).
		Find(&deployments).Error
}

func (s *DeploymentStore) GetForEnv(envId string) ([]store.Deployment, error) {
	var deployments []store.Deployment
	return deployments, s.preloadDeployment(s.db).
		Where(&store.Deployment{EnvId: envId}).
		Find(&deployments).Error
}

func (s *DeploymentStore) GetForCell(cellId string) ([]store.Deployment, error) {
	var deployments []store.Deployment
	cell := store.Cell{Common: store.Common{Id: cellId}}
	err := s.preloadDeployment(s.db).
		Model(&cell).
		Association("Deployments").
		Find(&deployments)
	return deployments, err
}

func (s *DeploymentStore) DeleteDeployment(appId string, envId string, id uint) error {
	return s.db.Delete(&store.Deployment{Id: id, AppId: appId, EnvId: envId}).Error
}
