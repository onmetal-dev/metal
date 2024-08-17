package dbstore

import (
	"github.com/onmetal-dev/metal/lib/store"
	"go.jetify.com/typeid"
	"gorm.io/gorm"
)

type ServerStore struct {
	db *gorm.DB
}

var _ store.ServerStore = &ServerStore{}

type NewServerStoreParams struct {
	DB *gorm.DB
}

func NewServerStore(params NewServerStoreParams) *ServerStore {
	return &ServerStore{
		db: params.DB,
	}
}

func (s *ServerStore) Create(server store.Server) (store.Server, error) {
	tid, _ := typeid.WithPrefix("server")
	server.Id = tid.String()
	return server, s.db.Create(&server).Error
}

func (s *ServerStore) Get(id string) (store.Server, error) {
	var server store.Server
	return server, s.db.Where("id = ?", id).First(&server).Error
}

func (s *ServerStore) UpdateServerStatus(serverId string, status store.ServerStatus) error {
	return s.db.Model(&store.Server{}).Where("id = ?", serverId).Update("status", status).Error
}

func (s *ServerStore) UpdateServerPublicIpv4(serverId string, publicIpv4 string) error {
	return s.db.Model(&store.Server{}).Where("id = ?", serverId).Update("public_ipv4", &publicIpv4).Error
}

func (s *ServerStore) UpdateProviderId(serverId string, providerId string) error {
	return s.db.Model(&store.Server{}).Where("id = ?", serverId).Update("provider_id", &providerId).Error
}

func (s *ServerStore) GetServersForTeam(teamId string) ([]store.Server, error) {
	var servers []store.Server
	return servers, s.db.Where("team_id = ?", teamId).Find(&servers).Error
}
