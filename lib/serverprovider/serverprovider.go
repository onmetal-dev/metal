// Package serverprovider provides a way to interact with a server provider, e.g. Hetzner.
// It abstracts away the details of the provider like ordering, provisioning, getting available offerings, etc.
// It can be thought of as a generic API across all providers.
// All state is kept at the server provider itself--this means all IDs are server-provider specific and no state is shared across providers and no state is kept within Metal.
package serverprovider

import "errors"

type Offering struct {
	Id          string
	Name        string
	Description string
	Locations   []string
	Prices      []Price
	Addons      []Addon
}

type Price struct {
	Currency      string
	Location      string
	AmountMonthly float64
	AmountSetup   float64
}

type Addon struct {
	Id          string
	Name        string
	Description string
	Min         int
	Max         int
	Prices      []Price
}

type Order struct {
	OfferingId string
	LocationId string
	AddonIds   []string
}

type Transaction struct {
	Id            string
	Status        TransactionStatus
	StatusDetails string
	ServerId      string
	OfferingId    string
	Location      string
	AddonIds      []string
}

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusCanceled  TransactionStatus = "canceled"
)

var ErrTransactionNotFound = errors.New("transaction not found")

type Server struct {
	Id            string
	ProviderSlug  string
	OfferingId    string
	Location      string
	Status        ServerStatus
	StatusDetails string
	Ipv4          string
	Ipv6          string
}

type ServerStatus string

const (
	ServerStatusPending ServerStatus = "pending"
	ServerStatusRunning ServerStatus = "running"
)

var ErrServerNotFound = errors.New("server not found")

type ServerProvider interface {
	Slug() string
	GetCurrentOfferings() ([]Offering, error)
	OrderServer(order Order) (Transaction, error)
	GetTransaction(id string) (Transaction, error)
	GetServer(serverId string) (Server, error)
}
