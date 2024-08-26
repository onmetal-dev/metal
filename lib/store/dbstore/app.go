package dbstore

import (
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/onmetal-dev/metal/lib/validate"
	"go.jetify.com/typeid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AppStore struct {
	db *gorm.DB
}

var _ store.AppStore = &AppStore{}

type NewAppStoreParams struct {
	DB *gorm.DB
}

func NewAppStore(params NewAppStoreParams) *AppStore {
	return &AppStore{
		db: params.DB,
	}
}

func (s *AppStore) Create(opts store.CreateAppOptions) (store.App, error) {
	tid, _ := typeid.WithPrefix("app")
	if err := validate.Struct(opts); err != nil {
		return store.App{}, err
	}
	app := store.App{
		Common: store.Common{
			Id: tid.String(),
		},
		TeamId: opts.TeamId,
		UserId: opts.UserId,
		Name:   opts.Name,
	}
	return app, s.db.Create(&app).Error
}

func (s *AppStore) Get(id string) (store.App, error) {
	var app store.App
	return app, s.db.Where("id = ?", id).First(&app).Error
}

func (s *AppStore) GetForTeam(teamId string) ([]store.App, error) {
	var apps []store.App
	return apps, s.db.Where("team_id = ?", teamId).Find(&apps).Error
}

func (s *AppStore) CreateAppSettings(opts store.CreateAppSettingsOptions) (store.AppSettings, error) {
	tid, _ := typeid.WithPrefix("appsettings")
	appSettings := store.AppSettings{
		Common: store.Common{
			Id: tid.String(),
		},
		TeamId:        opts.TeamId,
		AppId:         opts.AppId,
		Ports:         datatypes.NewJSONType(opts.Ports),
		ExternalPorts: datatypes.NewJSONType(opts.ExternalPorts),
		Resources:     datatypes.NewJSONType(opts.Resources),
	}
	return appSettings, s.db.Create(&appSettings).Error
}

func (s *AppStore) GetAppSettings(id string) (store.AppSettings, error) {
	var appSettings store.AppSettings
	return appSettings, s.db.Where("id = ?", id).First(&appSettings).Error
}
