package store

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type WaitlistedUser struct {
	Email     string    `gorm:"primaryKey" validate:"required,email"`
	CreatedAt time.Time `gorm:"index"`
}

type WaitlistStore interface {
	Add(email string) error
}

type InvitedUser struct {
	Email     string    `gorm:"primaryKey" validate:"required,email"`
	CreatedAt time.Time `gorm:"index"`
}

type InviteStore interface {
	Add(email string) error
	Get(email string) (*InvitedUser, error)
}

type User struct {
	Id              string       `gorm:"primaryKey" json:"id"`
	Email           string       `json:"email"`
	Password        string       `json:"-"`
	TeamMemberships []TeamMember `json:"team_memberships"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`

	// relations
	Servers []Server `gorm:"foreignKey:UserId"`
}

type UserStore interface {
	CreateUser(email string, password string) error
	GetUser(email string) (*User, error)
	GetUserById(id string) (*User, error)
}

type Team struct {
	Id               string             `gorm:"primaryKey" json:"id"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	StripeCustomerId string             `json:"stripe_customer_id"`
	AgePublicKey     string             `json:"age_public_key"`
	AgePrivateKey    string             `json:"-"`
	InvitedMembers   []TeamMemberInvite `json:"invited_members"`
	Members          []TeamMember       `json:"members"`
	PaymentMethods   []PaymentMethod    `json:"payment_methods"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`

	// relations
	Servers []Server `gorm:"foreignKey:TeamId"`
	Cells   []Cell   `gorm:"foreignKey:TeamId"`
}

type PaymentMethod struct {
	Id                    string    `gorm:"primaryKey" json:"id"`
	TeamId                string    `json:"team_id" gorm:"uniqueIndex:idx_team_payment_method"`
	StripePaymentMethodId string    `json:"stripe_payment_method_id" gorm:"uniqueIndex:idx_team_payment_method"`
	Default               bool      `json:"default"`
	Type                  string    `json:"type"` // e.g., "card", "bank_account"
	Last4                 string    `json:"last4"`
	ExpirationMonth       int       `json:"expiration_month,omitempty"`
	ExpirationYear        int       `json:"expiration_year,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type TeamRole string

const (
	TeamRoleAdmin  TeamRole = "admin"
	TeamRoleMember TeamRole = "member"
)

type TeamMemberInvite struct {
	TeamId    string
	Email     string    `json:"email" gorm:"index"`
	Role      TeamRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamMember struct {
	UserId    string    `gorm:"primaryKey" json:"user_id"`
	TeamId    string    `gorm:"primaryKey" json:"team_id"`
	Role      TeamRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TeamStore interface {
	CreateTeam(name string, description string) (*Team, error)
	GetTeam(id string) (*Team, error)
	AddUserToTeam(userId string, teamId string) error
	RemoveUserFromTeam(userId string, teamId string) error
	CreateTeamInvite(email string, teamId string) error
	DeleteTeamInvite(email string, teamId string) error
	GetInvitesForEmail(email string) ([]TeamMemberInvite, error)
	CreateStripeCustomer(teamId string, billingEmail string) error
	AddPaymentMethod(teamId string, paymentMethodData PaymentMethod) error
	RemovePaymentMethod(teamId string, paymentMethodId string) error
	GetPaymentMethods(teamId string) ([]PaymentMethod, error)
}

// Location of a server.
// Servers have a location that can be broken down into
// - continent, e.g. North America
// - country, e.g. United States of America
// - city, e.g. Dallas
type Location struct {
	Id        string // e.g. hel1
	Continent string
	Country   string
	City      string
}

// Cpu description on a server offering
type Cpu struct {
	Brand        string
	Family       string
	Name         string
	Cores        int
	Threads      int
	BaseSpeedGHz float64
	MaxSpeedGHz  float64
	Arch         string
}

// Disk available on a server
type Disk struct {
	SizeGB int
	Type   string // e.g. ssd, hdd
	Device string // e.g. sda, sdb
}

type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

// Price of a server offering
type Price struct {
	LocationId               string
	Setup                    float64
	Hourly                   float64
	Daily                    float64
	Monthly                  float64
	MonthlyOneYearUpfront    *float64 // potentially discounted
	MonthlyThreeYearsUpfront *float64 // potentially discounted
	Currency                 Currency
}

// Bandwidth available on a server
type Bandwidth struct {
	SpeedGbps int
	Unlimited bool
}

type ProviderSlug string

const (
	ProviderSlugHetzner ProviderSlug = "hetzner"
	ProviderSlugOVHUS   ProviderSlug = "ovh-us"
)

// ServerOffering is a server that can be purchased.
type ServerOffering struct {
	Id             string
	ProviderSlug   ProviderSlug // e.g. hetzner
	Type           string       // e.g. AX42
	Description    string
	Locations      []Location
	Prices         []Price
	Cpu            Cpu
	MemoryGB       int
	Disks          []Disk
	TotalStorageGB int
	Bandwidth      Bandwidth
}

type ServerOfferingStore interface {
	GetServerOfferings() ([]ServerOffering, error)
	GetServerOffering(id string) (*ServerOffering, error)
}

type ServerStatus string

const (
	ServerStatusPendingPayment  ServerStatus = "pending-payment"
	ServerStatusPendingProvider ServerStatus = "pending-provider"
	ServerStatusRunning         ServerStatus = "running"
	ServerStatusRunningCanceled ServerStatus = "running-canceled"
	ServerStatusDestroyed       ServerStatus = "destroyed"
)

type Server struct {
	// Id is our unique id for the server
	Id string
	// TeamId is the team that owns this server
	TeamId string
	// UserId is the user that created this server
	UserId string
	// OfferingId is the id of the server offering that this server is based on
	OfferingId string
	// LocationId is the location of the server
	LocationId string
	// Status is the current status of the server
	Status ServerStatus

	// ProviderSlug is the provider name, e.g. hetzner
	ProviderSlug string
	// ProviderId is the id within the provider
	ProviderId *string
	// PublicIpv4 is the public IP of the server
	PublicIpv4 *string

	// CellId is the id of the cell that this server belongs to
	CellId *string `gorm:"constraint:OnDelete:SET NULL"`

	// BillingStripeHourlyUsage keeps track of hourly billing for the server (nullabe in case of different billing schemes in the future)
	BillingStripeUsageBasedHourly *ServerBillingStripeUsageBasedHourly

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// ServerStripeUsageBasedHourly keeps track of the last time we recorded a usage event for a server
type ServerBillingStripeUsageBasedHourly struct {
	ServerId string `gorm:"primaryKey"`
	// EventName is the event we send to stripe. We configure this as a "count" i.e. 1 event = 1 hour of usage
	EventName string
	// LastEventSent is the last time we sent an event to stripe and can be used to determine if we need to send one or more new events
	LastEventSent sql.NullTime
}

type ServerStore interface {
	Create(s Server) (Server, error)
	Get(id string) (Server, error)
	UpdateServerStatus(serverId string, status ServerStatus) error
	UpdateServerPublicIpv4(serverId string, publicIpv4 string) error
	UpdateProviderId(serverId string, providerId string) error
	GetServersForTeam(teamId string) ([]Server, error)
}

type CellType string

const (
	CellTypeTalos CellType = "talos"
)

// Cell is a group of servers, primarily used as to separate workloads from any combination of environments / teams / etc.
type Cell struct {
	Id          string `gorm:"primaryKey"`
	Type        CellType
	Name        string
	TeamId      string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// relations
	Servers       []Server
	TalosCellData *TalosCellData `gorm:"foreignKey:CellId;references:Id"`
}

type TalosCellData struct {
	CellId      string `gorm:"primaryKey"`
	Talosconfig string
	Kubecfg     string
	Config      []byte
}

type CellStore interface {
	Create(c Cell) (Cell, error)
	Get(id string) (Cell, error)
	GetForTeam(teamId string) ([]Cell, error)
	UpdateTalosCellData(cellId string, talosCellData TalosCellData) error
	AddServer(cellId string, server Server) error
}
