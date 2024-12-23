package dbstore

import (
	"context"

	"filippo.io/age"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/customer"
	"go.jetify.com/typeid"
	"gorm.io/gorm"
)

type TeamStore struct {
	db             *gorm.DB
	stripeCustomer *customer.Client
}

var _ store.TeamStore = TeamStore{}

type NewTeamStoreParams struct {
	DB             *gorm.DB
	StripeCustomer *customer.Client
}

func NewTeamStore(params NewTeamStoreParams) *TeamStore {
	return &TeamStore{
		db:             params.DB,
		stripeCustomer: params.StripeCustomer,
	}
}

func (s TeamStore) CreateTeam(name string, description string) (*store.Team, error) {
	tid, _ := typeid.WithPrefix("team")
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, err
	}
	team := store.Team{
		Common: store.Common{
			Id: tid.String(),
		},
		Name:           name,
		Description:    description,
		AgePublicKey:   identity.Recipient().String(),
		AgePrivateKey:  identity.String(),
		Members:        []store.TeamMember{},
		InvitedMembers: []store.TeamMemberInvite{},
	}
	if err := s.db.Create(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

func (s *TeamStore) preloadTeam(query *gorm.DB) *gorm.DB {
	return query.Preload("Members").Preload("Members.User").Preload("InvitedMembers").Preload("PaymentMethods").Preload("Cells").Preload("Envs").Preload("Apps")
}

func (s TeamStore) GetTeam(ctx context.Context, id string) (*store.Team, error) {
	var team store.Team
	if err := s.preloadTeam(s.db).WithContext(ctx).Where("id = ?", id).First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

func (s TeamStore) GetTeamKeys(id string) (string, string, error) {
	var team store.Team
	err := s.db.Where("id = ?", id).First(&team).Error
	if err != nil {
		return "", "", err
	}
	return team.AgePublicKey, team.AgePrivateKey, nil
}

func (s TeamStore) AddUserToTeam(userId string, teamId string) error {
	return s.db.Create(&store.TeamMember{
		UserId: userId,
		TeamId: teamId,
		Role:   store.TeamRoleMember,
	}).Error
}

func (s TeamStore) RemoveUserFromTeam(userId string, teamId string) error {
	return s.db.Where("user_id = ? AND team_id = ?", userId, teamId).Delete(&store.TeamMember{}).Error
}

func (s TeamStore) CreateTeamInvite(email string, teamId string) error {
	return s.db.Create(&store.TeamMemberInvite{
		TeamId: teamId,
		Email:  email,
		Role:   store.TeamRoleMember,
	}).Error
}

func (s TeamStore) DeleteTeamInvite(email string, teamId string) error {
	return s.db.Where("email = ? AND team_id = ?", email, teamId).Delete(&store.TeamMemberInvite{}).Error
}

func (s TeamStore) GetInvitesForEmail(email string) ([]store.TeamMemberInvite, error) {
	var invites []store.TeamMemberInvite
	err := s.db.Where("email = ?", email).Find(&invites).Error
	return invites, err
}

func (s TeamStore) CreateStripeCustomer(ctx context.Context, teamId string, billingEmail string) error {
	team, err := s.GetTeam(ctx, teamId)
	if err != nil {
		return err
	}

	params := &stripe.CustomerParams{
		Description: stripe.String(team.Name),
		Email:       stripe.String(billingEmail),
	}
	cust, err := s.stripeCustomer.New(params)
	if err != nil {
		return err
	}

	team.StripeCustomerId = cust.ID
	return s.db.Save(team).Error
}

func (s TeamStore) AddPaymentMethod(ctx context.Context, teamId string, paymentMethodData store.PaymentMethod) error {
	tid, _ := typeid.WithPrefix("pm")
	paymentMethodData.Id = tid.String()
	paymentMethodData.TeamId = teamId
	if err := s.db.Create(&paymentMethodData).Error; err != nil {
		if err == gorm.ErrDuplicatedKey {
			// we're ok with this, the payment method already exists
			return nil
		}
		return err
	}

	team, err := s.GetTeam(ctx, teamId)
	if err != nil {
		return err
	}
	if paymentMethodData.Default {
		if _, err := s.stripeCustomer.Update(team.StripeCustomerId, &stripe.CustomerParams{
			InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
				DefaultPaymentMethod: stripe.String(paymentMethodData.StripePaymentMethodId),
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s TeamStore) RemovePaymentMethod(teamId string, paymentMethodId string) error {
	return s.db.Where("team_id = ? AND id = ?", teamId, paymentMethodId).Delete(&store.PaymentMethod{}).Error
}

func (s TeamStore) GetPaymentMethods(teamId string) ([]store.PaymentMethod, error) {
	var paymentMethods []store.PaymentMethod
	err := s.db.Where("team_id = ?", teamId).Find(&paymentMethods).Error
	return paymentMethods, err
}
