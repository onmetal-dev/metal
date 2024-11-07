package dbstore

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/onmetal-dev/metal/lib/store"
	"go.jetify.com/typeid"
	"gorm.io/gorm"
)

type ApiTokenStore struct {
	db *gorm.DB
}

var _ store.ApiTokenStore = &ApiTokenStore{}

func NewApiTokenStore(db *gorm.DB) *ApiTokenStore {
	return &ApiTokenStore{db: db}
}

func (s *ApiTokenStore) Create(teamId string, creatorId string, name string, scope store.ApiTokenScope) (*store.ApiToken, error) {
	tid, _ := typeid.WithPrefix("apitoken")
	token := &store.ApiToken{
		Common: store.Common{
			Id: tid.String(),
		},
		TeamId:    teamId,
		CreatorId: creatorId,
		Name:      name,
		Token:     generateUniqueToken(), // Implement this function to generate a unique token
		Scope:     scope,
	}
	if err := s.db.Create(token).Error; err != nil {
		return nil, err
	}
	return token, nil
}

func (s *ApiTokenStore) Get(id string) (*store.ApiToken, error) {
	var token store.ApiToken
	if err := s.db.Where("id = ?", id).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *ApiTokenStore) GetByToken(token string) (*store.ApiToken, error) {
	var apiToken store.ApiToken
	if err := s.db.Where("token = ?", token).First(&apiToken).Error; err != nil {
		return nil, err
	}
	return &apiToken, nil
}

func (s *ApiTokenStore) List(teamId string) ([]store.ApiToken, error) {
	var tokens []store.ApiToken
	if err := s.db.Where("team_id = ?", teamId).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *ApiTokenStore) Delete(id string) error {
	return s.db.Where("id = ?", id).Delete(&store.ApiToken{}).Error
}

func (s *ApiTokenStore) UpdateLastUsedAt(id string, lastUsedAt time.Time) error {
	return s.db.Model(&store.ApiToken{}).Where("id = ?", id).Update("last_used_at", lastUsedAt).Error
}

// generateUniqueToken generates a unique token
func generateUniqueToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const tokenLength = 100
	token := make([]byte, tokenLength)
	_, err := rand.Read(token)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate random bytes: %v", err))
	}
	for i := range token {
		token[i] = charset[int(token[i])%len(charset)]
	}
	return string(token)
}
