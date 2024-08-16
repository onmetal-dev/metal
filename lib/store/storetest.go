package store

import (
	"testing"
)

type TestStoresConfig struct {
	UserStore UserStore
	TeamStore TeamStore
}

func NewStoreTestSuite(stores TestStoresConfig) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("User and Team Operations", func(t *testing.T) {
			// Create a user
			email := "test@example.com"
			password := "password123"
			err := stores.UserStore.CreateUser(email, password)
			if err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}

			// Get the created user
			user, err := stores.UserStore.GetUser(email)
			if err != nil {
				t.Fatalf("Failed to get user: %v", err)
			}
			if user.Email != email {
				t.Errorf("Expected user email %s, got %s", email, user.Email)
			}

			// Create a team
			teamName := "Test Team"
			teamDesc := "A team for testing"
			teamCreated, err := stores.TeamStore.CreateTeam(teamName, teamDesc)
			if err != nil {
				t.Fatalf("Failed to create team: %v", err)
			}

			// Get the created team (assuming the first team created has ID "1")
			team, err := stores.TeamStore.GetTeam(teamCreated.ID)
			if err != nil {
				t.Fatalf("Failed to get team: %v", err)
			}
			if team.Name != teamName {
				t.Errorf("Expected team name %s, got %s", teamName, team.Name)
			}

			// Add user to team
			err = stores.TeamStore.AddUserToTeam(user.ID, team.ID)
			if err != nil {
				t.Fatalf("Failed to add user to team: %v", err)
			}

			// Verify user is part of the team
			updatedTeam, err := stores.TeamStore.GetTeam(team.ID)
			if err != nil {
				t.Fatalf("Failed to get updated team: %v", err)
			}
			if len(updatedTeam.Members) != 1 || updatedTeam.Members[0].UserID != user.ID {
				t.Errorf("User not found in team members")
			}

			// Create a team invite
			inviteEmail := "invite@example.com"
			err = stores.TeamStore.CreateTeamInvite(inviteEmail, team.ID)
			if err != nil {
				t.Fatalf("Failed to create team invite: %v", err)
			}

			// Verify invite is present in the team
			teamWithInvite, err := stores.TeamStore.GetTeam(team.ID)
			if err != nil {
				t.Fatalf("Failed to get team with invite: %v", err)
			}
			if len(teamWithInvite.InvitedMembers) != 1 || teamWithInvite.InvitedMembers[0].Email != inviteEmail {
				t.Errorf("Invite not found in team invited members")
			}

			// Verify invite is returned by GetInvitesForEmail
			invites, err := stores.TeamStore.GetInvitesForEmail(inviteEmail)
			if err != nil {
				t.Fatalf("Failed to get invites for email: %v", err)
			}
			if len(invites) != 1 || invites[0].Email != inviteEmail {
				t.Errorf("Invite not found in GetInvitesForEmail result")
			}

			// Delete the team invite
			err = stores.TeamStore.DeleteTeamInvite(inviteEmail, team.ID)
			if err != nil {
				t.Fatalf("Failed to delete team invite: %v", err)
			}

			// Verify invite has been deleted
			teamAfterDelete, err := stores.TeamStore.GetTeam(team.ID)
			if err != nil {
				t.Fatalf("Failed to get team after invite deletion: %v", err)
			}
			if len(teamAfterDelete.InvitedMembers) != 0 {
				t.Errorf("Invite still present in team after deletion")
			}

			invitesAfterDelete, err := stores.TeamStore.GetInvitesForEmail(inviteEmail)
			if err != nil {
				t.Fatalf("Failed to get invites after deletion: %v", err)
			}
			if len(invitesAfterDelete) != 0 {
				t.Errorf("Invite still present in GetInvitesForEmail result after deletion")
			}
		})
	}
}
