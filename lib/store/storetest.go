package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestStoresConfig struct {
	UserStore   UserStore
	TeamStore   TeamStore
	ServerStore ServerStore
	CellStore   CellStore
}

func createUser(t *testing.T, stores TestStoresConfig, email, password string) User {
	require := require.New(t)
	err := stores.UserStore.CreateUser(email, password)
	require.NoError(err, "Failed to create user")
	user, err := stores.UserStore.GetUser(email)
	require.NotEmpty(user.Id, "Expected user id to be present")
	require.NoError(err, "Failed to get user")
	require.Equal(email, user.Email, "Expected user email %s, got %s", email, user.Email)
	return *user
}

func createTeam(t *testing.T, stores TestStoresConfig, name, description string) Team {
	require := require.New(t)
	team, err := stores.TeamStore.CreateTeam(name, description)
	require.NoError(err, "Failed to create team")
	require.NotEmpty(team.Id, "Expected team id to be present")
	require.Equal(name, team.Name, "Expected team name %s, got %s", name, team.Name)
	require.NotNil(team.Members, "Expected team members to be present")
	require.Equal(0, len(team.Members), "Expected team members to be empty")
	require.NotNil(team.InvitedMembers, "Expected team invited members to be present")
	require.Equal(0, len(team.InvitedMembers), "Expected team invited members to be empty")

	return *team
}

func createServer(t *testing.T, serverStore ServerStore, server Server) Server {
	require := require.New(t)
	s, err := serverStore.Create(server)
	require.NoError(err, "Failed to create team")
	require.Nil(s.CellId, "Expected no cell assignment on create")
	return s
}

func createCell(t *testing.T, cellStore CellStore, cell Cell) Cell {
	require := require.New(t)
	c, err := cellStore.Create(cell)
	require.NoError(err, "Failed to create cell")
	require.NotEmpty(c.Id, "Expected cell id to be present")
	return c
}

func addUserToTeam(t *testing.T, stores TestStoresConfig, userId, teamId string) {
	require := require.New(t)
	err := stores.TeamStore.AddUserToTeam(userId, teamId)
	require.NoError(err, "Failed to add user to team")

	updatedTeam, err := stores.TeamStore.GetTeam(teamId)
	require.NoError(err, "Failed to get updated team")
	require.Equal(1, len(updatedTeam.Members), "Expected team members to be 1")
	require.Equal(userId, updatedTeam.Members[0].UserId, "Expected user to be in team members")
}

func NewStoreTestSuite(stores TestStoresConfig) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("User and Team Operations", func(t *testing.T) {
			require := require.New(t)
			email := "test@example.com"
			password := "password123"
			user := createUser(t, stores, email, password)

			teamName := "Test Team"
			teamDesc := "A team for testing"
			team := createTeam(t, stores, teamName, teamDesc)

			addUserToTeam(t, stores, user.Id, team.Id)

			// Create a team invite
			inviteEmail := "invite@example.com"
			err := stores.TeamStore.CreateTeamInvite(inviteEmail, team.Id)
			require.NoError(err, "Failed to create team invite")

			// Verify invite is present in the team
			teamWithInvite, err := stores.TeamStore.GetTeam(team.Id)
			require.NoError(err, "Failed to get team with invite")
			require.Equal(1, len(teamWithInvite.InvitedMembers), "Expected team invited members to be 1")
			require.Equal(inviteEmail, teamWithInvite.InvitedMembers[0].Email, "Expected invite email to be %s, got %s", inviteEmail, teamWithInvite.InvitedMembers[0].Email)

			// Verify invite is returned by GetInvitesForEmail
			invites, err := stores.TeamStore.GetInvitesForEmail(inviteEmail)
			require.NoError(err, "Failed to get invites for email")
			require.Equal(1, len(invites), "Expected invites to be 1")
			require.Equal(inviteEmail, invites[0].Email, "Expected invite email to be %s, got %s", inviteEmail, invites[0].Email)

			// Delete the team invite
			err = stores.TeamStore.DeleteTeamInvite(inviteEmail, team.Id)
			require.NoError(err, "Failed to delete team invite")

			// Verify invite has been deleted
			teamAfterDelete, err := stores.TeamStore.GetTeam(team.Id)
			require.NoError(err, "Failed to get team after invite deletion")
			require.Equal(0, len(teamAfterDelete.InvitedMembers), "Expected team invited members to be empty")

			invitesAfterDelete, err := stores.TeamStore.GetInvitesForEmail(inviteEmail)
			require.NoError(err, "Failed to get invites after deletion")
			require.Equal(0, len(invitesAfterDelete), "Expected invites to be empty")
		})
		t.Run("Server and cell operations", func(t *testing.T) {
			require := require.New(t)
			email := "test@example.com"
			password := "password123"
			user := createUser(t, stores, email, password)

			teamName := "Test Team"
			teamDesc := "A team for testing"
			team := createTeam(t, stores, teamName, teamDesc)

			addUserToTeam(t, stores, user.Id, team.Id)

			server := createServer(t, stores.ServerStore, Server{
				TeamId:       team.Id,
				UserId:       user.Id,
				OfferingId:   "AX102",
				LocationId:   "HEL1",
				ProviderSlug: "hetzner",
				Status:       ServerStatusPendingProvider,
			})
			require.Equal(team.Id, server.TeamId, "Expected server team id to be %s, got %s", team.Id, server.TeamId)
			require.NotEmpty(server.Id, "Expected server id to be present")

			cell := createCell(t, stores.CellStore, Cell{
				Name:    "default",
				TeamId:  team.Id,
				Servers: []Server{server},
				TalosCellData: &TalosCellData{
					Talosconfig: "test",
					Config:      []byte("test"),
				},
			})
			require.NotEmpty(cell.Id, "Expected cell id to be present")
			require.Equal(server.Id, cell.Servers[0].Id, "Expected cell servers to be %s, got %s", server.Id, cell.Servers[0].Id)

			getCell, err := stores.CellStore.Get(cell.Id)
			require.NoError(err, "Failed to get cell")
			require.Equal(1, len(getCell.Servers), "Expected cell servers to be 1")
			require.Equal(server.Id, getCell.Servers[0].Id, "Expected cell servers to be %s, got %s", server.Id, getCell.Servers[0].Id)
			require.NotNil(getCell.TalosCellData, "Expected cell talos data to be present")
			require.Equal("test", getCell.TalosCellData.Talosconfig, "Expected cell talos config to be %s, got %s", "test", getCell.TalosCellData.Talosconfig)
			require.Equal([]byte("test"), getCell.TalosCellData.Config, "Expected cell talos config to be %s, got %s", []byte("test"), getCell.TalosCellData.Config)

		})

	}
}
