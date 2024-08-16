package serverprovider

import (
	"errors"
	"fmt"

	"github.com/vultr/govultr/v3"
)

type Vultr struct {
	client *govultr.Client
}

var _ ServerProvider = &Vultr{}

type VultrOption func(*Vultr) error

func WithVultrClient(client *govultr.Client) VultrOption {
	return func(v *Vultr) error {
		if client == nil {
			return errors.New("Vultr client cannot be nil")
		}
		v.client = client
		return nil
	}
}

func NewVultr(opts ...VultrOption) (*Vultr, error) {
	v := &Vultr{}
	for _, opt := range opts {
		if err := opt(v); err != nil {
			return nil, err
		}
	}
	if v.client == nil {
		return nil, errors.New("Vultr client is required")
	}
	return v, nil
}

const VultrSlug = "vultr"

func (v *Vultr) Slug() string {
	return VultrSlug
}

func (v *Vultr) GetCurrentOfferings() ([]Offering, error) {
	// TODO: Implement this method
	return []Offering{}, fmt.Errorf("GetCurrentOfferings not implemented")
}

func (v *Vultr) OrderServer(order Order) (Transaction, error) {
	// TODO: Implement this method
	return Transaction{}, fmt.Errorf("OrderServer not implemented")
}

func (v *Vultr) GetTransaction(id string) (Transaction, error) {
	// TODO: Implement this method
	return Transaction{}, fmt.Errorf("GetTransaction not implemented")
}

func (v *Vultr) GetServer(serverId string) (Server, error) {
	// TODO: Implement this method
	return Server{}, fmt.Errorf("GetServer not implemented")
}
