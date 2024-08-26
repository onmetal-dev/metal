package dbstore

import (
	"errors"

	"github.com/onmetal-dev/metal/cmd/app/hash"
	"github.com/onmetal-dev/metal/lib/store"
	"go.jetify.com/typeid"
	"gorm.io/gorm"
)

type UserStore struct {
	db           *gorm.DB
	passwordhash hash.PasswordHash
}

type NewUserStoreParams struct {
	DB           *gorm.DB
	PasswordHash hash.PasswordHash
}

func NewUserStore(params NewUserStoreParams) *UserStore {
	return &UserStore{
		db:           params.DB,
		passwordhash: params.PasswordHash,
	}
}

func (s *UserStore) CreateUser(email string, password string) error {
	hashedPassword, err := s.passwordhash.GenerateFromPassword(password)
	if err != nil {
		return err
	}

	tid, _ := typeid.WithPrefix("user")
	return s.db.Create(&store.User{
		Common: store.Common{
			Id: tid.String(),
		},
		Email:    email,
		Password: hashedPassword,
	}).Error
}

func (s *UserStore) GetUser(email string) (*store.User, error) {
	var user store.User
	if err := s.db.Preload("TeamMemberships").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) GetUserById(id string) (*store.User, error) {
	var user store.User
	if err := s.db.Preload("TeamMemberships").Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
