package serverprovider

import (
	"errors"
	"fmt"

	"github.com/ovh/go-ovh/ovh"
)

type OVH struct {
	client *ovh.Client
}

var _ ServerProvider = &OVH{}

type OVHOption func(*OVH) error

func WithOVHClient(client *ovh.Client) OVHOption {
	return func(o *OVH) error {
		if client == nil {
			return errors.New("OVH client cannot be nil")
		}
		o.client = client
		return nil
	}
}

func NewOVH(opts ...OVHOption) (*OVH, error) {
	o := &OVH{}
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	if o.client == nil {
		return nil, errors.New("OVH client is required")
	}
	return o, nil
}

const OVHSlug = "ovh"

func (o *OVH) Slug() string {
	return OVHSlug
}

func (o *OVH) GetCurrentOfferings() ([]Offering, error) {
	// TODO: Implement this method
	return []Offering{}, fmt.Errorf("GetCurrentOfferings not implemented")
}

func (o *OVH) OrderServer(order Order) (Transaction, error) {
	// TODO: Implement this method
	return Transaction{}, fmt.Errorf("OrderServer not implemented")
}

func (o *OVH) GetTransaction(id string) (Transaction, error) {
	// TODO: Implement this method
	return Transaction{}, fmt.Errorf("GetTransaction not implemented")
}

func (o *OVH) GetServer(serverId string) (Server, error) {
	// TODO: Implement this method
	return Server{}, fmt.Errorf("GetServer not implemented")
}
