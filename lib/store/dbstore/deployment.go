package dbstore

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"

	"filippo.io/age"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/onmetal-dev/metal/lib/validate"
	"github.com/samber/lo"
	"go.jetify.com/typeid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/hints"
)

type DeploymentStore struct {
	db          *gorm.DB
	getTeamKeys func(id string) (string, string, error)
}

var _ store.DeploymentStore = &DeploymentStore{}

type NewDeploymentStoreParams struct {
	DB          *gorm.DB
	GetTeamKeys func(id string) (string, string, error)
}

func NewDeploymentStore(params NewDeploymentStoreParams) (*DeploymentStore, error) {
	if params.DB == nil {
		return nil, fmt.Errorf("db is required")
	}
	if params.GetTeamKeys == nil {
		return nil, fmt.Errorf("getTeamKeys is required")
	}
	return &DeploymentStore{
		db:          params.DB,
		getTeamKeys: params.GetTeamKeys,
	}, nil
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
	public, _, err := s.getTeamKeys(opts.TeamId)
	if err != nil {
		return store.AppEnvVars{}, err
	}
	recipient, err := age.ParseX25519Recipient(public)
	if err != nil {
		return store.AppEnvVars{}, err
	}
	encryptedEnvVars := make([]store.EnvVar, len(opts.EnvVars))
	for i, envVar := range opts.EnvVars {
		encryptedValue, err := ageEncryptValue(envVar.Value, recipient)
		if err != nil {
			return store.AppEnvVars{}, err
		}
		encryptedEnvVars[i] = store.EnvVar{
			Name:  envVar.Name,
			Value: encryptedValue,
		}
	}
	tid, _ := typeid.WithPrefix("appenvvars")
	appEnvVars := store.AppEnvVars{
		Common:  store.Common{Id: tid.String()},
		TeamId:  opts.TeamId,
		EnvId:   opts.EnvId,
		AppId:   opts.AppId,
		EnvVars: datatypes.NewJSONType(encryptedEnvVars),
	}
	return appEnvVars, s.db.Create(&appEnvVars).Error
}

func ageEncryptValue(value string, recipient *age.X25519Recipient) (string, error) {
	out := &bytes.Buffer{}
	w, err := age.Encrypt(out, recipient)
	if err != nil {
		return "", err
	}
	if _, err := io.WriteString(w, value); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(out.Bytes()), nil
}

func ageDecryptValue(encryptedValue string, identity *age.X25519Identity) (string, error) {
	decodedValue, err := base64.StdEncoding.DecodeString(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted value: %v", err)
	}

	r, err := age.Decrypt(bytes.NewReader(decodedValue), identity)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt value: %v", err)
	}

	decryptedValue := &bytes.Buffer{}
	if _, err := io.Copy(decryptedValue, r); err != nil {
		return "", fmt.Errorf("failed to read decrypted value: %v", err)
	}

	return decryptedValue.String(), nil
}

func (s *DeploymentStore) decryptAppEnvVars(appEnvVars *store.AppEnvVars) error {
	_, private, err := s.getTeamKeys(appEnvVars.TeamId)
	if err != nil {
		return err
	}
	identity, err := age.ParseX25519Identity(private)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	decryptedEnvVars := make([]store.EnvVar, len(appEnvVars.EnvVars.Data()))
	for i, envVar := range appEnvVars.EnvVars.Data() {
		decryptedValue, err := ageDecryptValue(envVar.Value, identity)
		if err != nil {
			return fmt.Errorf("failed to decrypt env var: %v", err)
		}
		decryptedEnvVars[i] = store.EnvVar{
			Name:  envVar.Name,
			Value: decryptedValue,
		}
	}

	appEnvVars.EnvVars = datatypes.NewJSONType(decryptedEnvVars)
	return nil
}

func (s *DeploymentStore) GetAppEnvVars(id string) (store.AppEnvVars, error) {
	var appEnvVars store.AppEnvVars
	if err := s.db.First(&appEnvVars, "id = ?", id).Error; err != nil {
		return store.AppEnvVars{}, err
	}

	if err := s.decryptAppEnvVars(&appEnvVars); err != nil {
		return store.AppEnvVars{}, err
	}

	return appEnvVars, nil
}

func (s *DeploymentStore) GetAppEnvVarsForAppEnv(appId string, envId string) ([]store.AppEnvVars, error) {
	var appEnvVars []store.AppEnvVars
	if err := s.db.Where(&store.AppEnvVars{AppId: appId, EnvId: envId}).Find(&appEnvVars).Error; err != nil {
		return nil, err
	}

	for i := range appEnvVars {
		if err := s.decryptAppEnvVars(&appEnvVars[i]); err != nil {
			return nil, err
		}
	}

	return appEnvVars, nil
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
		Status:        store.DeploymentStatusPending,
		Replicas:      opts.Replicas,
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
	if err := s.preloadDeployment(s.db).First(&deployment).Error; err != nil {
		return store.Deployment{}, err
	}
	if err := s.decryptAppEnvVars(&deployment.AppEnvVars); err != nil {
		return store.Deployment{}, err
	}
	return deployment, nil
}

func (s *DeploymentStore) GetForTeam(teamId string) ([]store.Deployment, error) {
	var deployments []store.Deployment
	err := s.preloadDeployment(s.db).
		Where(&store.Deployment{TeamId: teamId}).
		Find(&deployments).Error
	if err != nil {
		return nil, err
	}
	for i := range deployments {
		if err := s.decryptAppEnvVars(&deployments[i].AppEnvVars); err != nil {
			return nil, err
		}
	}
	return deployments, nil
}

func (s *DeploymentStore) GetForApp(appId string) ([]store.Deployment, error) {
	var deployments []store.Deployment
	err := s.preloadDeployment(s.db).
		Clauses(hints.UseIndex("idx_app_createdat")).
		Where(&store.Deployment{AppId: appId}).
		Order("created_at DESC").
		Find(&deployments).Error
	if err != nil {
		return nil, err
	}
	for i := range deployments {
		if err := s.decryptAppEnvVars(&deployments[i].AppEnvVars); err != nil {
			return nil, err
		}
	}
	return deployments, nil
}

func (s *DeploymentStore) GetForEnv(envId string) ([]store.Deployment, error) {
	var deployments []store.Deployment
	err := s.preloadDeployment(s.db).
		Where(&store.Deployment{EnvId: envId}).
		Find(&deployments).Error
	if err != nil {
		return nil, err
	}
	for i := range deployments {
		if err := s.decryptAppEnvVars(&deployments[i].AppEnvVars); err != nil {
			return nil, err
		}
	}
	return deployments, nil
}

func (s *DeploymentStore) GetForCell(cellId string) ([]store.Deployment, error) {
	var deployments []store.Deployment
	cell := store.Cell{Common: store.Common{Id: cellId}}
	err := s.preloadDeployment(s.db).
		Model(&cell).
		Association("Deployments").
		Find(&deployments)
	if err != nil {
		return nil, err
	}
	for i := range deployments {
		if err := s.decryptAppEnvVars(&deployments[i].AppEnvVars); err != nil {
			return nil, err
		}
	}
	return deployments, nil
}

func (s *DeploymentStore) DeleteDeployment(appId string, envId string, id uint) error {
	return s.db.Delete(&store.Deployment{Id: id, AppId: appId, EnvId: envId}).Error
}

func (s *DeploymentStore) UpdateDeploymentStatus(appId string, envId string, id uint, status store.DeploymentStatus, statusReason string) error {
	return s.db.Where(&store.Deployment{AppId: appId, EnvId: envId, Id: id}).
		Select("Status", "StatusReason").
		Updates(store.Deployment{Status: status, StatusReason: statusReason}).Error
}
