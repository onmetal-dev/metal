package dbstore

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/onmetal-dev/metal/lib/store"
	"gorm.io/gorm"
)

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

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

func (s *WaitlistStore) Add(email string) error {
	u := &store.WaitlistedUser{
		Email: email,
	}
	if err := validate.Struct(u); err != nil {
		return fmt.Errorf("not a valid email: '%s'", email)
	}

	// post to loops
	data := url.Values{}
	data.Set("email", email)
	data.Set("userGroup", "waitlist")
	req, err := http.NewRequest("POST", s.loopsWaitlistFormUrl, strings.NewReader(data.Encode()))
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

	return s.db.Create(&store.WaitlistedUser{
		Email: email,
	}).Error
}
