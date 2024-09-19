package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Common struct {
	Id        string         `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// ErrDuplicateWaitlistEntry is returned when trying to add a duplicate email to the waitlist
var ErrDuplicateWaitlistEntry = errors.New("email already exists in waitlist")

type WaitlistedUser struct {
	Email     string    `gorm:"primaryKey" validate:"required,email"`
	CreatedAt time.Time `gorm:"index"`
}

type WaitlistStore interface {
	Add(ctx context.Context, email string) error
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
	Common
	Email           string       `json:"email"`
	Password        string       `json:"-"`
	TeamMemberships []TeamMember `json:"team_memberships"`

	// relations
	Servers []Server `gorm:"foreignKey:UserId"`
	Apps    []App    `gorm:"foreignKey:UserId"`
}

type UserStore interface {
	CreateUser(email string, password string) error
	GetUser(email string) (*User, error)
	GetUserById(id string) (*User, error)
}

type Team struct {
	Common
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	StripeCustomerId string             `json:"stripe_customer_id"`
	AgePublicKey     string             `json:"age_public_key"`
	AgePrivateKey    string             `json:"-"`
	InvitedMembers   []TeamMemberInvite `json:"invited_members"`
	Members          []TeamMember       `json:"members"`
	PaymentMethods   []PaymentMethod    `json:"payment_methods"`

	// relations
	Servers     []Server     `gorm:"foreignKey:TeamId"`
	Cells       []Cell       `gorm:"foreignKey:TeamId"`
	Apps        []App        `gorm:"foreignKey:TeamId"`
	Envs        []Env        `gorm:"foreignKey:TeamId"`
	Deployments []Deployment `gorm:"foreignKey:TeamId"`
	ApiTokens   []ApiToken   `gorm:"foreignKey:TeamId"`
}

type PaymentMethod struct {
	Common
	TeamId                string `json:"team_id" gorm:"uniqueIndex:idx_team_payment_method"`
	StripePaymentMethodId string `json:"stripe_payment_method_id" gorm:"uniqueIndex:idx_team_payment_method"`
	Default               bool   `json:"default"`
	Type                  string `json:"type"` // e.g., "card", "bank_account"
	Last4                 string `json:"last4"`
	ExpirationMonth       int    `json:"expiration_month,omitempty"`
	ExpirationYear        int    `json:"expiration_year,omitempty"`
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
	User      User      `gorm:"foreignKey:UserId"`
	TeamId    string    `gorm:"primaryKey" json:"team_id"`
	Role      TeamRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TeamStore interface {
	CreateTeam(name string, description string) (*Team, error)
	GetTeam(ctx context.Context, id string) (*Team, error)
	GetTeamKeys(id string) (string, string, error)
	AddUserToTeam(userId string, teamId string) error
	RemoveUserFromTeam(userId string, teamId string) error
	CreateTeamInvite(email string, teamId string) error
	DeleteTeamInvite(email string, teamId string) error
	GetInvitesForEmail(email string) ([]TeamMemberInvite, error)
	CreateStripeCustomer(ctx context.Context, teamId string, billingEmail string) error
	AddPaymentMethod(ctx context.Context, teamId string, paymentMethodData PaymentMethod) error
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
	Common
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
	BillingStripeUsageBasedHourly *ServerBillingStripeUsageBasedHourly `gorm:"foreignKey:ServerId;references:Id"`
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
	UpdateServerBillingStripeUsageBasedHourly(serverId string, usageBasedHourly *ServerBillingStripeUsageBasedHourly) error
	UpdateProviderId(serverId string, providerId string) error
	GetServersForTeam(ctx context.Context, teamId string) ([]Server, error)
}

type CellType string

const (
	CellTypeTalos CellType = "talos"
)

// Cell is a group of servers, primarily used as to separate workloads from any combination of environments / teams / etc.
type Cell struct {
	Common
	Type        CellType
	Name        string
	TeamId      string
	Description string

	// relations
	Servers       []Server
	TalosCellData *TalosCellData `gorm:"foreignKey:CellId;references:Id"`
	Deployments   []Deployment   `gorm:"many2many:deployment_cells;"`
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
	GetForTeam(ctx context.Context, teamId string) ([]Cell, error)
	UpdateTalosCellData(talosCellData *TalosCellData) error
	AddServer(cellId string, server Server) error
}

type App struct {
	Common
	TeamId    string    `json:"team_id" gorm:"index:idx_team_createdat"`
	UserId    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `gorm:"index:idx_team_createdat"`
}

type Artifact struct {
	Image Image `json:"image"`
}

type Image struct {
	Name string `json:"name"`
}
type Port struct {
	Name  string `validate:"required,lowercasealphanumhyphen"`
	Port  int    `validate:"required"`
	Proto string `validate:"required,oneof=http"`
}

type Ports []Port

type ExternalPort struct {
	Name     string `validate:"required,lowercasealphanumhyphen"`
	PortName string `validate:"required,lowercasealphanumhyphen"` // reference to a Port
	Proto    string `validate:"required,oneof=http https"`
	Port     int    `validate:"required"`
}

type ExternalPorts []ExternalPort

type Resources struct {
	Limits   ResourceLimits   `json:"limits"`
	Requests ResourceRequests `json:"requests"`
}

type ResourceLimits struct {
	CpuCores  float64 `json:"cpu_cores"`
	MemoryMiB int     `json:"memory_mib"`
}

type ResourceRequests struct {
	CpuCores  float64 `json:"cpu_cores"`
	MemoryMiB int     `json:"memory_mib"`
}

type AppSettings struct {
	Common
	TeamId        string                            `json:"team_id"`
	AppId         string                            `json:"app_id"`
	Artifact      datatypes.JSONType[Artifact]      `gorm:"type:jsonb" json:"artifact"`
	Ports         datatypes.JSONType[Ports]         `gorm:"type:jsonb" json:"ports"`
	ExternalPorts datatypes.JSONType[ExternalPorts] `gorm:"type:jsonb" json:"external_ports"`
	Resources     datatypes.JSONType[Resources]     `gorm:"type:jsonb" json:"resources"`
}

type CreateAppOptions struct {
	Name   string `validate:"required,lowercasealphanumhyphen"`
	TeamId string `validate:"required"`
	UserId string `validate:"required"`
}

type CreateAppSettingsOptions struct {
	TeamId        string        `validate:"required"`
	AppId         string        `validate:"required"`
	Artifact      Artifact      `validate:"required"`
	Ports         Ports         `validate:"required"`
	ExternalPorts ExternalPorts `validate:"required"`
	Resources     Resources     `validate:"required"`
}

var ErrAppNotFound = errors.New("app not found")

type AppStore interface {
	Create(opts CreateAppOptions) (App, error)
	Get(ctx context.Context, id string) (App, error)
	Delete(ctx context.Context, id string) error
	GetForTeam(ctx context.Context, teamId string) ([]App, error)
	CreateAppSettings(opts CreateAppSettingsOptions) (AppSettings, error)
	GetAppSettings(id string) (AppSettings, error)
}

// in order to deploy we need a concept of environments, as well as environment variables
// environment variables are tied to an evironment + app. Similar to AppSettings, they are immutable. We mint a new environment variables object when changes are made.
// This effectifely snapshots the environment variables, allowing for rollbacks.

type Env struct {
	Common
	TeamId string
	Name   string
}

type EnvVar struct {
	Name  string
	Value string
}

type AppEnvVars struct {
	Common
	TeamId  string
	EnvId   string
	AppId   string
	EnvVars datatypes.JSONType[[]EnvVar]
}

type DeploymentType string

const (
	DeploymentTypeDeploy   DeploymentType = "deploy"
	DeploymentTypeRollback DeploymentType = "rollback"
	DeploymentTypeScale    DeploymentType = "scale"
	DeploymentTypeRestart  DeploymentType = "restart"
)

type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusDeploying DeploymentStatus = "deploying"
	DeploymentStatusRunning   DeploymentStatus = "running"
	DeploymentStatusFailed    DeploymentStatus = "failed"
	DeploymentStatusStopped   DeploymentStatus = "stopped"
)

// Deployment has a monotonic id that is incremented for each deployment of an app/env combination
type Deployment struct {
	Id            uint   `gorm:"primarykey"`
	EnvId         string `gorm:"primaryKey;index"`
	AppId         string `gorm:"primaryKey;index:idx_app_createdat"`
	TeamId        string `gorm:"index:idx_team_createdat"`
	Env           Env    `gorm:"foreignKey:EnvId"`
	App           App    `gorm:"foreignKey:AppId"`
	Type          DeploymentType
	Status        DeploymentStatus
	StatusReason  string
	Replicas      int
	AppSettingsId string
	AppSettings   AppSettings `gorm:"foreignKey:AppSettingsId"`
	AppEnvVarsId  string
	AppEnvVars    AppEnvVars `gorm:"foreignKey:AppEnvVarsId"`
	Cells         []Cell     `gorm:"many2many:deployment_cells;"`
	CreatedAt     time.Time  `gorm:"index:idx_app_createdat;index:idx_team_createdat"`
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (d *Deployment) BeforeCreate(tx *gorm.DB) error {
	var maxID uint
	result := tx.Model(&Deployment{}).
		Where("env_id = ? AND app_id = ?", d.EnvId, d.AppId).
		Select("COALESCE(MAX(id), 0)").
		Scan(&maxID)

	if result.Error != nil {
		return result.Error
	}
	d.Id = maxID + 1
	return nil
}

type CreateEnvOptions struct {
	TeamId string `validate:"required"`
	Name   string `validate:"required,lowercasealphanumhyphen"`
}

type CreateAppEnvVarOptions struct {
	TeamId  string   `validate:"required"`
	EnvId   string   `validate:"required"`
	AppId   string   `validate:"required"`
	EnvVars []EnvVar `validate:"required"`
}

type CreateDeploymentOptions struct {
	TeamId        string         `validate:"required"`
	EnvId         string         `validate:"required"`
	AppId         string         `validate:"required"`
	Type          DeploymentType `validate:"required,oneof=deploy rollback scale restart"`
	AppSettingsId string         `validate:"required"`
	AppEnvVarsId  string         `validate:"required"`
	CellIds       []string       `validate:"required"`
	Replicas      int            `validate:"required"`
}

// DeploymentStore allows for
// - creating, retrieving (by teamId), and deleting environments
// - creating, retrieving (by teamId, appId, envId), and deleting AppEnvVars
// - creating, retrieving (by teamId or by Id or by appId, or by envId, or by cellId), and deleting Deployments
type DeploymentStore interface {
	CreateEnv(opts CreateEnvOptions) (Env, error)
	GetEnv(id string) (Env, error)
	GetEnvsForTeam(teamId string) ([]Env, error)
	DeleteEnv(id string) error

	CreateAppEnvVars(opts CreateAppEnvVarOptions) (AppEnvVars, error)
	GetAppEnvVars(id string) (AppEnvVars, error)
	GetAppEnvVarsForAppEnv(appId string, envId string) ([]AppEnvVars, error)
	DeleteAppEnvVars(id string) error

	Create(opts CreateDeploymentOptions) (Deployment, error)
	Get(appId string, envId string, id uint) (Deployment, error)
	GetForTeam(ctx context.Context, teamId string) ([]Deployment, error)
	GetForApp(ctx context.Context, appId string) ([]Deployment, error)
	GetForEnv(envId string) ([]Deployment, error)
	GetForCell(cellId string) ([]Deployment, error)
	DeleteDeployment(appId string, envId string, id uint) error
	UpdateDeploymentStatus(appId string, envId string, id uint, status DeploymentStatus, statusReason string) error
}

// ApiTokenScope represents the access level of an API token
type ApiTokenScope string

const (
	ApiTokenScopeAdmin ApiTokenScope = "admin"
)

// ApiToken represents an API token for a team
type ApiToken struct {
	Common
	TeamId     string        `json:"team_id" gorm:"index"`
	CreatorId  string        `json:"creator_id"`
	Name       string        `json:"name"`
	Token      string        `json:"token" gorm:"uniqueIndex"`
	Scope      ApiTokenScope `json:"scope"`
	LastUsedAt *time.Time    `json:"last_used"`
}

// ApiTokenStore interface for managing API tokens
type ApiTokenStore interface {
	Create(teamId string, creatorId string, name string, scope ApiTokenScope) (*ApiToken, error)
	Get(id string) (*ApiToken, error)
	GetByToken(token string) (*ApiToken, error)
	List(teamId string) ([]ApiToken, error)
	Delete(id string) error
	UpdateLastUsedAt(id string, lastUsedAt time.Time) error
}
