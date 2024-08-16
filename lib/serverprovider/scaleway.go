package serverprovider

import (
	"errors"
	"fmt"

	instance "github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type Scaleway struct {
	client *scw.Client
	api    *instance.API
}

var _ ServerProvider = &Scaleway{}

type ScalewayOption func(*Scaleway) error

func WithScalewayClient(client *scw.Client) ScalewayOption {
	return func(s *Scaleway) error {
		if client == nil {
			return errors.New("Scaleway client cannot be nil")
		}
		s.client = client
		s.api = instance.NewAPI(client)
		return nil
	}
}

func NewScaleway(opts ...ScalewayOption) (*Scaleway, error) {
	s := &Scaleway{}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	if s.client == nil {
		return nil, errors.New("Scaleway client is required")
	}
	return s, nil
}

const ScalewaySlug = "scaleway"

func (s *Scaleway) Slug() string {
	return ScalewaySlug
}

func (s *Scaleway) GetCurrentOfferings() ([]Offering, error) {
	// TODO: Implement this method
	return []Offering{}, fmt.Errorf("GetCurrentOfferings not implemented")
}

func (s *Scaleway) OrderServer(order Order) (Transaction, error) {
	// TODO: Implement this method
	return Transaction{}, fmt.Errorf("OrderServer not implemented")
}

func (s *Scaleway) GetTransaction(id string) (Transaction, error) {
	// TODO: Implement this method
	return Transaction{}, fmt.Errorf("GetTransaction not implemented")
}

func (s *Scaleway) GetServer(serverId string) (Server, error) {
	// TODO: Implement this method
	return Server{}, fmt.Errorf("GetServer not implemented")
}
