package dbstore

import (
	"fmt"

	"github.com/onmetal-dev/metal/lib/store"
	"gorm.io/gorm"
)

type InviteStore struct {
	db *gorm.DB
}

type NewInviteStoreParams struct {
	DB *gorm.DB
}

func NewInviteStore(params NewInviteStoreParams) *InviteStore {
	return &InviteStore{
		db: params.DB,
	}
}

func (s *InviteStore) Add(email string) error {
	u := &store.InvitedUser{
		Email: email,
	}
	if err := validate.Struct(u); err != nil {
		return fmt.Errorf("not a valid email: '%s'", email)
	}
	return s.db.Create(u).Error
}

func (s *InviteStore) Get(email string) (*store.InvitedUser, error) {
	var user store.InvitedUser
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
