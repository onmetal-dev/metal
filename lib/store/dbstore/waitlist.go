package dbstore

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/onmetal-dev/metal/lib/validate"
	"gorm.io/gorm"
)

type WaitlistStore struct {
	loopsWaitlistFormUrl string
	db                   *gorm.DB
}

type NewWaitlistStoreParams struct {
	LoopsWaitlistFormUrl string
	DB                   *gorm.DB
}

func NewWaitlistStore(params NewWaitlistStoreParams) *WaitlistStore {
	return &WaitlistStore{
		loopsWaitlistFormUrl: params.LoopsWaitlistFormUrl,
		db:                   params.DB,
	}
}

func (s *WaitlistStore) postToLoops(ctx context.Context, email string) error {
	data := url.Values{}
	data.Set("email", email)
	data.Set("userGroup", "waitlist")
	req, err := http.NewRequestWithContext(ctx, "POST", s.loopsWaitlistFormUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to post to loops, status code: %d", resp.StatusCode)
	}
	return nil
}

func (s *WaitlistStore) Add(ctx context.Context, email string) error {
	u := &store.WaitlistedUser{
		Email: email,
	}
	if err := validate.Struct(u); err != nil {
		return fmt.Errorf("not a valid email: '%s'", email)
	}

	if s.loopsWaitlistFormUrl != "" {
		if err := s.postToLoops(ctx, email); err != nil {
			// just log the error, but don't fail the function
			// we can resolve the inconsistency between the db and loops out of band
			logger.FromContext(ctx).Error("failed to post to loops", "error", err)
		}
	}

	if err := s.db.Create(&store.WaitlistedUser{
		Email: email,
	}).Error; err != nil {
		if err == gorm.ErrDuplicatedKey {
			return store.ErrDuplicateWaitlistEntry
		}
		return err
	}

	return nil
}
