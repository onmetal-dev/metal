package serverprovider

import (
	"errors"
	"fmt"

	metal "github.com/equinix/equinix-sdk-go/services/metalv1"
)

type Equinix struct {
	client *metal.APIClient
}

var _ ServerProvider = &Equinix{}

type EquinixOption func(*Equinix) error

func WithEquinixClient(client *metal.APIClient) EquinixOption {
	return func(e *Equinix) error {
		if client == nil {
			return errors.New("Equinix client cannot be nil")
		}
		e.client = client
		return nil
	}
}

func NewEquinix(opts ...EquinixOption) (*Equinix, error) {
	e := &Equinix{}
	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}
	if e.client == nil {
		return nil, errors.New("Equinix client is required")
	}
	return e, nil
}

const EquinixSlug = "equinix"

func (e *Equinix) Slug() string {
	return EquinixSlug
}

func (e *Equinix) GetCurrentOfferings() ([]Offering, error) {
	// TODO: Implement this method
	return []Offering{}, fmt.Errorf("GetCurrentOfferings not implemented")
}

func (e *Equinix) OrderServer(order Order) (Transaction, error) {
	// TODO: Implement this method
	return Transaction{}, fmt.Errorf("OrderServer not implemented")
}

func (e *Equinix) GetTransaction(id string) (Transaction, error) {
	// TODO: Implement this method
	return Transaction{}, fmt.Errorf("GetTransaction not implemented")
}

func (e *Equinix) GetServer(serverId string) (Server, error) {
	// TODO: Implement this method
	return Server{}, fmt.Errorf("GetServer not implemented")
}
