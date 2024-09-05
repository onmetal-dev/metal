package mock

import (
	"context"

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
