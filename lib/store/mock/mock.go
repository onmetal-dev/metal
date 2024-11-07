package mock

import (
	"context"
	"time"

	"github.com/onmetal-dev/metal/lib/store"

	"github.com/stretchr/testify/mock"
)

type UserStoreMock struct {
	mock.Mock
}

var _ store.UserStore = &UserStoreMock{}

func (m *UserStoreMock) CreateUser(email string, password string) error {
	args := m.Called(email, password)

	return args.Error(0)
}

func (m *UserStoreMock) GetUser(email string) (*store.User, error) {
	args := m.Called(email)
	return args.Get(0).(*store.User), args.Error(1)
}

func (m *UserStoreMock) GetUserById(id string) (*store.User, error) {
	args := m.Called(id)
	return args.Get(0).(*store.User), args.Error(1)
}

type InviteStoreMock struct {
	mock.Mock
}

var _ store.InviteStore = &InviteStoreMock{}

func (m *InviteStoreMock) Add(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *InviteStoreMock) Get(email string) (*store.InvitedUser, error) {
	args := m.Called(email)
	return args.Get(0).(*store.InvitedUser), args.Error(1)
}

type TeamStoreMock struct {
	mock.Mock
}

var _ store.TeamStore = &TeamStoreMock{}

func (m *TeamStoreMock) CreateTeam(name string, description string) (*store.Team, error) {
	args := m.Called(name, description)
	return args.Get(0).(*store.Team), args.Error(1)
}

func (m *TeamStoreMock) GetTeam(ctx context.Context, id string) (*store.Team, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*store.Team), args.Error(1)
}

func (m *TeamStoreMock) GetTeamKeys(id string) (string, string, error) {
	args := m.Called(id)
	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

func (m *TeamStoreMock) AddUserToTeam(userId string, teamId string) error {
	args := m.Called(userId, teamId)
	return args.Error(0)
}

func (m *TeamStoreMock) RemoveUserFromTeam(userId string, teamId string) error {
	args := m.Called(userId, teamId)
	return args.Error(0)
}

func (m *TeamStoreMock) CreateTeamInvite(email string, teamId string) error {
	args := m.Called(email, teamId)
	return args.Error(0)
}

func (m *TeamStoreMock) DeleteTeamInvite(email string, teamId string) error {
	args := m.Called(email, teamId)
	return args.Error(0)
}

func (m *TeamStoreMock) GetInvitesForEmail(email string) ([]store.TeamMemberInvite, error) {
	args := m.Called(email)
	return args.Get(0).([]store.TeamMemberInvite), args.Error(1)
}

func (m *TeamStoreMock) CreateStripeCustomer(ctx context.Context, teamId string, billingEmail string) error {
	args := m.Called(ctx, teamId, billingEmail)
	return args.Error(0)
}

func (m *TeamStoreMock) AddPaymentMethod(ctx context.Context, teamId string, paymentMethodData store.PaymentMethod) error {
	args := m.Called(ctx, teamId, paymentMethodData)
	return args.Error(0)
}

func (m *TeamStoreMock) RemovePaymentMethod(teamId string, paymentMethodId string) error {
	args := m.Called(teamId, paymentMethodId)
	return args.Error(0)
}

func (m *TeamStoreMock) GetPaymentMethods(teamId string) ([]store.PaymentMethod, error) {
	args := m.Called(teamId)
	return args.Get(0).([]store.PaymentMethod), args.Error(1)
}

type AppStoreMock struct {
	mock.Mock
}

var _ store.AppStore = &AppStoreMock{}

func (m *AppStoreMock) Create(opts store.CreateAppOptions) (store.App, error) {
	args := m.Called(opts)
	return args.Get(0).(store.App), args.Error(1)
}

func (m *AppStoreMock) Get(ctx context.Context, id string) (store.App, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.App), args.Error(1)
}

func (m *AppStoreMock) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *AppStoreMock) GetForTeam(ctx context.Context, teamId string) ([]store.App, error) {
	args := m.Called(ctx, teamId)
	return args.Get(0).([]store.App), args.Error(1)
}

func (m *AppStoreMock) CreateAppSettings(opts store.CreateAppSettingsOptions) (store.AppSettings, error) {
	args := m.Called(opts)
	return args.Get(0).(store.AppSettings), args.Error(1)
}

func (m *AppStoreMock) GetAppSettings(id string) (store.AppSettings, error) {
	args := m.Called(id)
	return args.Get(0).(store.AppSettings), args.Error(1)
}

type DeploymentStoreMock struct {
	mock.Mock
}

var _ store.DeploymentStore = &DeploymentStoreMock{}

func (m *DeploymentStoreMock) CreateEnv(opts store.CreateEnvOptions) (store.Env, error) {
	args := m.Called(opts)
	return args.Get(0).(store.Env), args.Error(1)
}

func (m *DeploymentStoreMock) GetEnv(id string) (store.Env, error) {
	args := m.Called(id)
	return args.Get(0).(store.Env), args.Error(1)
}

func (m *DeploymentStoreMock) GetEnvsForTeam(teamId string) ([]store.Env, error) {
	args := m.Called(teamId)
	return args.Get(0).([]store.Env), args.Error(1)
}

func (m *DeploymentStoreMock) DeleteEnv(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *DeploymentStoreMock) CreateAppEnvVars(opts store.CreateAppEnvVarOptions) (store.AppEnvVars, error) {
	args := m.Called(opts)
	return args.Get(0).(store.AppEnvVars), args.Error(1)
}

func (m *DeploymentStoreMock) GetAppEnvVars(id string) (store.AppEnvVars, error) {
	args := m.Called(id)
	return args.Get(0).(store.AppEnvVars), args.Error(1)
}

func (m *DeploymentStoreMock) GetAppEnvVarsForAppEnv(appId string, envId string) ([]store.AppEnvVars, error) {
	args := m.Called(appId, envId)
	return args.Get(0).([]store.AppEnvVars), args.Error(1)
}

func (m *DeploymentStoreMock) DeleteAppEnvVars(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *DeploymentStoreMock) Create(opts store.CreateDeploymentOptions) (store.Deployment, error) {
	args := m.Called(opts)
	return args.Get(0).(store.Deployment), args.Error(1)
}

func (m *DeploymentStoreMock) Get(appId string, envId string, id uint) (store.Deployment, error) {
	args := m.Called(appId, envId, id)
	return args.Get(0).(store.Deployment), args.Error(1)
}

func (m *DeploymentStoreMock) GetForTeam(ctx context.Context, teamId string) ([]store.Deployment, error) {
	args := m.Called(ctx, teamId)
	return args.Get(0).([]store.Deployment), args.Error(1)
}

func (m *DeploymentStoreMock) GetForApp(ctx context.Context, appId string) ([]store.Deployment, error) {
	args := m.Called(ctx, appId)
	return args.Get(0).([]store.Deployment), args.Error(1)
}

func (m *DeploymentStoreMock) GetLatestForAppEnv(ctx context.Context, appId string, envId string) (*store.Deployment, error) {
	args := m.Called(ctx, appId, envId)
	return args.Get(0).(*store.Deployment), args.Error(1)
}

func (m *DeploymentStoreMock) GetForEnv(envId string) ([]store.Deployment, error) {
	args := m.Called(envId)
	return args.Get(0).([]store.Deployment), args.Error(1)
}

func (m *DeploymentStoreMock) GetForCell(cellId string) ([]store.Deployment, error) {
	args := m.Called(cellId)
	return args.Get(0).([]store.Deployment), args.Error(1)
}

func (m *DeploymentStoreMock) DeleteDeployment(appId string, envId string, id uint) error {
	args := m.Called(appId, envId, id)
	return args.Error(0)
}

func (m *DeploymentStoreMock) UpdateDeploymentStatus(appId string, envId string, id uint, status store.DeploymentStatus, statusReason string) error {
	args := m.Called(appId, envId, id, status, statusReason)
	return args.Error(0)
}

type ApiTokenStoreMock struct {
	mock.Mock
}

var _ store.ApiTokenStore = &ApiTokenStoreMock{}

func (m *ApiTokenStoreMock) Create(teamId string, creatorId string, name string, scope store.ApiTokenScope) (*store.ApiToken, error) {
	args := m.Called(teamId, creatorId, name, scope)
	return args.Get(0).(*store.ApiToken), args.Error(1)
}

func (m *ApiTokenStoreMock) Get(id string) (*store.ApiToken, error) {
	args := m.Called(id)
	return args.Get(0).(*store.ApiToken), args.Error(1)
}

func (m *ApiTokenStoreMock) GetByToken(token string) (*store.ApiToken, error) {
	args := m.Called(token)
	return args.Get(0).(*store.ApiToken), args.Error(1)
}

func (m *ApiTokenStoreMock) List(teamId string) ([]store.ApiToken, error) {
	args := m.Called(teamId)
	return args.Get(0).([]store.ApiToken), args.Error(1)
}

func (m *ApiTokenStoreMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *ApiTokenStoreMock) UpdateLastUsedAt(id string, lastUsedAt time.Time) error {
	args := m.Called(id, lastUsedAt)
	return args.Error(0)
}

type BuildStoreMock struct {
	mock.Mock
}

var _ store.BuildStore = &BuildStoreMock{}

func (m *BuildStoreMock) Init(ctx context.Context, opts store.InitBuildOptions) (store.Build, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(store.Build), args.Error(1)
}

func (m *BuildStoreMock) Get(ctx context.Context, id string) (store.Build, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(store.Build), args.Error(1)
}

func (m *BuildStoreMock) UpdateStatus(ctx context.Context, id string, status store.BuildStatus, statusReason string) error {
	args := m.Called(ctx, id, status, statusReason)
	return args.Error(0)
}

func (m *BuildStoreMock) UpdateLogs(ctx context.Context, id string, logs store.BuildLogs) error {
	args := m.Called(ctx, id, logs)
	return args.Error(0)
}

func (m *BuildStoreMock) UpdateArtifacts(ctx context.Context, id string, artifacts []store.Artifact) error {
	args := m.Called(ctx, id, artifacts)
	return args.Error(0)
}
