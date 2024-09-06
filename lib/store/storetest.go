package store

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
)

type TestStoresConfig struct {
	WaitlistStore   WaitlistStore
	UserStore       UserStore
	TeamStore       TeamStore
	ServerStore     ServerStore
	CellStore       CellStore
	AppStore        AppStore
	DeploymentStore DeploymentStore // Add this line
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

func addUserToTeam(t *testing.T, ctx context.Context, stores TestStoresConfig, userId, teamId string) {
	require := require.New(t)
	err := stores.TeamStore.AddUserToTeam(userId, teamId)
	require.NoError(err, "Failed to add user to team")

	updatedTeam, err := stores.TeamStore.GetTeam(ctx, teamId)
	require.NoError(err, "Failed to get updated team")
	require.Equal(1, len(updatedTeam.Members), "Expected team members to be 1")
	require.Equal(userId, updatedTeam.Members[0].UserId, "Expected user to be in team members")
}

func NewStoreTestSuite(stores TestStoresConfig) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("User and Team Operations", func(t *testing.T) {
			ctx := context.Background()
			require := require.New(t)
			email := "test@example.com"
			password := "password123"
			user := createUser(t, stores, email, password)

			teamName := "Test Team"
			teamDesc := "A team for testing"
			team := createTeam(t, stores, teamName, teamDesc)

			addUserToTeam(t, ctx, stores, user.Id, team.Id)

			// Create a team invite
			inviteEmail := "invite@example.com"
			err := stores.TeamStore.CreateTeamInvite(inviteEmail, team.Id)
			require.NoError(err, "Failed to create team invite")

			// Verify invite is present in the team
			teamWithInvite, err := stores.TeamStore.GetTeam(ctx, team.Id)
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
			teamAfterDelete, err := stores.TeamStore.GetTeam(ctx, team.Id)
			require.NoError(err, "Failed to get team after invite deletion")
			require.Equal(0, len(teamAfterDelete.InvitedMembers), "Expected team invited members to be empty")

			invitesAfterDelete, err := stores.TeamStore.GetInvitesForEmail(inviteEmail)
			require.NoError(err, "Failed to get invites after deletion")
			require.Equal(0, len(invitesAfterDelete), "Expected invites to be empty")
		})
		t.Run("Server and cell operations", func(t *testing.T) {
			require := require.New(t)
			ctx := context.Background()
			email := "test@example.com"
			password := "password123"
			user := createUser(t, stores, email, password)

			teamName := "Test Team"
			teamDesc := "A team for testing"
			team := createTeam(t, stores, teamName, teamDesc)

			addUserToTeam(t, ctx, stores, user.Id, team.Id)

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

			err = stores.CellStore.UpdateTalosCellData(&TalosCellData{
				CellId:      getCell.Id,
				Talosconfig: "test2",
			})
			require.NoError(err, "Failed to update cell")

			getCell, err = stores.CellStore.Get(cell.Id)
			require.NoError(err, "Failed to get cell")
			require.Equal("test2", getCell.TalosCellData.Talosconfig, "Expected cell talos config to be %s, got %s", "test2", getCell.TalosCellData.Talosconfig)

		})

		t.Run("App and AppSettings Operations", func(t *testing.T) {
			require := require.New(t)
			ctx := context.Background()

			// Create a user and team for the app
			email := "apptest@example.com"
			password := "password123"
			user := createUser(t, stores, email, password)

			teamName := "App Test Team"
			teamDesc := "A team for testing apps"
			team := createTeam(t, stores, teamName, teamDesc)

			addUserToTeam(t, ctx, stores, user.Id, team.Id)

			// Create an app
			appName := "Test App"
			createAppOpts := CreateAppOptions{
				Name:   appName,
				TeamId: team.Id,
				UserId: user.Id,
			}
			app, err := stores.AppStore.Create(createAppOpts)
			require.NoError(err, "Failed to create app")
			require.NotEmpty(app.Id, "Expected app id to be present")
			require.Equal(appName, app.Name, "Expected app name to match")
			require.Equal(team.Id, app.TeamId, "Expected app team id to match")
			require.Equal(user.Id, app.UserId, "Expected app user id to match")

			// Get the created app
			fetchedApp, err := stores.AppStore.Get(ctx, app.Id)
			require.NoError(err, "Failed to get app")
			require.Equal(app.Id, fetchedApp.Id, "Expected fetched app id to match")
			require.Equal(app.Name, fetchedApp.Name, "Expected fetched app name to match")

			// Get apps for the team
			teamApps, err := stores.AppStore.GetForTeam(ctx, team.Id)
			require.NoError(err, "Failed to get apps for team")
			require.Equal(1, len(teamApps), "Expected one app for the team")
			require.Equal(app.Id, teamApps[0].Id, "Expected team app id to match")

			// Create app settings
			ports := Ports{
				{Name: "http", Port: 80, Proto: "http"},
			}
			externalPorts := ExternalPorts{
				{Name: "web", PortName: "http", Proto: "https", Port: 443},
			}
			resources := Resources{
				Limits: ResourceLimits{
					CpuCores:  1,
					MemoryMiB: 1024,
				},
				Requests: ResourceRequests{
					CpuCores:  0.5,
					MemoryMiB: 512,
				},
			}
			createAppSettingsOpts := CreateAppSettingsOptions{
				TeamId:        team.Id,
				AppId:         app.Id,
				Ports:         ports,
				ExternalPorts: externalPorts,
				Resources:     resources,
			}
			appSettings, err := stores.AppStore.CreateAppSettings(createAppSettingsOpts)
			require.NoError(err, "Failed to create app settings")
			require.NotEmpty(appSettings.Id, "Expected app settings id to be present")
			require.Equal(app.Id, appSettings.AppId, "Expected app settings app id to match")
			require.Equal(team.Id, appSettings.TeamId, "Expected app settings team id to match")

			// Get the created app settings
			fetchedAppSettings, err := stores.AppStore.GetAppSettings(appSettings.Id)
			require.NoError(err, "Failed to get app settings")
			require.Equal(appSettings.Id, fetchedAppSettings.Id, "Expected fetched app settings id to match")
			require.Equal(len(ports), len(fetchedAppSettings.Ports.Data()), "Expected fetched app settings ports to match")
			require.Equal(len(externalPorts), len(fetchedAppSettings.ExternalPorts.Data()), "Expected fetched app settings external ports to match")
			require.Equal(resources.Limits.CpuCores, fetchedAppSettings.Resources.Data().Limits.CpuCores, "Expected fetched app settings CPU limit to match")
			require.Equal(resources.Limits.MemoryMiB, fetchedAppSettings.Resources.Data().Limits.MemoryMiB, "Expected fetched app settings memory limit to match")
		})

		t.Run("Deployment Operations", func(t *testing.T) {
			require := require.New(t)
			ctx := context.Background()

			// Create a user and team for the deployment tests
			email := "deploytest@example.com"
			password := "password123"
			user := createUser(t, stores, email, password)

			teamName := "Deploy Test Team"
			teamDesc := "A team for testing deployments"
			team := createTeam(t, stores, teamName, teamDesc)

			addUserToTeam(t, ctx, stores, user.Id, team.Id)

			// Test Env operations
			t.Run("Env Operations", func(t *testing.T) {
				// Create Env
				createEnvOpts := CreateEnvOptions{
					TeamId: team.Id,
					Name:   "test-env",
				}
				env, err := stores.DeploymentStore.CreateEnv(createEnvOpts)
				require.NoError(err, "Failed to create env")
				require.NotEmpty(env.Id, "Expected env id to be present")
				require.Equal(createEnvOpts.Name, env.Name, "Expected env name to match")

				// Get Env
				fetchedEnv, err := stores.DeploymentStore.GetEnv(env.Id)
				require.NoError(err, "Failed to get env with ID %s", env.Id)
				require.Equal(env.Id, fetchedEnv.Id, "Expected fetched env id to match")
				require.Equal(env.Name, fetchedEnv.Name, "Expected fetched env name to match")

				// Get Envs for Team
				teamEnvs, err := stores.DeploymentStore.GetEnvsForTeam(team.Id)
				require.NoError(err, "Failed to get envs for team")
				require.Equal(1, len(teamEnvs), "Expected one env for the team")
				require.Equal(env.Id, teamEnvs[0].Id, "Expected team env id to match")

				// Delete Env
				err = stores.DeploymentStore.DeleteEnv(env.Id)
				require.NoError(err, "Failed to delete env")

				// Verify env is deleted
				_, err = stores.DeploymentStore.GetEnv(env.Id)
				require.Error(err, "Expected error when getting deleted env")
			})

			// Test AppEnvVars operations
			t.Run("AppEnvVars Operations", func(t *testing.T) {
				// Create App and Env for AppEnvVars
				app, _ := stores.AppStore.Create(CreateAppOptions{Name: "test-app", TeamId: team.Id, UserId: user.Id})
				env, _ := stores.DeploymentStore.CreateEnv(CreateEnvOptions{TeamId: team.Id, Name: "test-env"})

				// Create AppEnvVars
				createAppEnvVarsOpts := CreateAppEnvVarOptions{
					TeamId:  team.Id,
					EnvId:   env.Id,
					AppId:   app.Id,
					EnvVars: []EnvVar{{Name: "TEST_VAR", Value: "test_value"}},
				}
				appEnvVars, err := stores.DeploymentStore.CreateAppEnvVars(createAppEnvVarsOpts)
				require.NoError(err, "Failed to create app env vars")
				require.NotEmpty(appEnvVars.Id, "Expected app env vars id to be present")

				// Get AppEnvVars
				fetchedAppEnvVars, err := stores.DeploymentStore.GetAppEnvVars(appEnvVars.Id)
				require.NoError(err, "Failed to get app env vars")
				require.Equal(appEnvVars.Id, fetchedAppEnvVars.Id, "Expected fetched app env vars id to match")
				require.Equal(createAppEnvVarsOpts.EnvVars[0], fetchedAppEnvVars.EnvVars.Data()[0], "Expected fetched app env vars to match")

				// Get AppEnvVars for App and Env
				appEnvVarsList, err := stores.DeploymentStore.GetAppEnvVarsForAppEnv(app.Id, env.Id)
				require.NoError(err, "Failed to get app env vars for app and env")
				require.Equal(1, len(appEnvVarsList), "Expected one app env vars for the app and env")
				require.Equal(appEnvVars.Id, appEnvVarsList[0].Id, "Expected app env vars id to match")
				require.Equal(createAppEnvVarsOpts.EnvVars[0], appEnvVarsList[0].EnvVars.Data()[0], "Expected app env vars to match")

				// Delete AppEnvVars
				err = stores.DeploymentStore.DeleteAppEnvVars(appEnvVars.Id)
				require.NoError(err, "Failed to delete app env vars")

				// Verify app env vars is deleted
				_, err = stores.DeploymentStore.GetAppEnvVars(appEnvVars.Id)
				require.Error(err, "Expected error when getting deleted app env vars")
			})

			// Test Deployment operations
			t.Run("Deployment Operations", func(t *testing.T) {
				ctx := context.Background()
				app, _ := stores.AppStore.Create(CreateAppOptions{Name: "test-app", TeamId: team.Id, UserId: user.Id})
				env, _ := stores.DeploymentStore.CreateEnv(CreateEnvOptions{TeamId: team.Id, Name: "test-env"})
				appSettings, _ := stores.AppStore.CreateAppSettings(CreateAppSettingsOptions{
					TeamId:        team.Id,
					AppId:         app.Id,
					Ports:         Ports{{Name: "http", Port: 80, Proto: "http"}},
					ExternalPorts: ExternalPorts{{Name: "web", PortName: "http", Proto: "https", Port: 443}},
					Resources: Resources{
						Limits:   ResourceLimits{CpuCores: 1, MemoryMiB: 1024},
						Requests: ResourceRequests{CpuCores: 0.5, MemoryMiB: 512},
					},
				})
				appEnvVars, _ := stores.DeploymentStore.CreateAppEnvVars(CreateAppEnvVarOptions{
					TeamId:  team.Id,
					EnvId:   env.Id,
					AppId:   app.Id,
					EnvVars: []EnvVar{{Name: "TEST_VAR", Value: "test_value"}},
				})
				cell, _ := stores.CellStore.Create(Cell{TeamId: team.Id, Name: "test-cell"})

				// Create Deployment
				createDeploymentOpts := CreateDeploymentOptions{
					TeamId:        team.Id,
					EnvId:         env.Id,
					AppId:         app.Id,
					Type:          DeploymentTypeDeploy,
					AppSettingsId: appSettings.Id,
					AppEnvVarsId:  appEnvVars.Id,
					CellIds:       []string{cell.Id},
				}
				deployment, err := stores.DeploymentStore.Create(createDeploymentOpts)
				require.NoError(err, "Failed to create deployment")
				require.NotEmpty(deployment.Id, "Expected deployment id to be present")
				require.Equal(uint(1), deployment.Id, "Expected first deployment id to be 1")

				// Get Deployment
				fetchedDeployment, err := stores.DeploymentStore.Get(app.Id, env.Id, deployment.Id)
				require.NoError(err, "Failed to get deployment")
				require.Equal(deployment.Id, fetchedDeployment.Id, "Expected fetched deployment id to match")

				// Create another deployment for the same app/env
				deployment2, err := stores.DeploymentStore.Create(createDeploymentOpts)
				require.NoError(err, "Failed to create second deployment")
				require.Equal(uint(2), deployment2.Id, "Expected second deployment id to be 2")

				// Get Deployments for Team
				teamDeployments, err := stores.DeploymentStore.GetForTeam(ctx, team.Id)
				require.NoError(err, "Failed to get deployments for team")
				require.Equal(2, len(teamDeployments), "Expected two deployments for the team")

				// Get Deployments for App
				appDeployments, err := stores.DeploymentStore.GetForApp(ctx, app.Id)
				require.NoError(err, "Failed to get deployments for app")
				require.Equal(2, len(appDeployments), "Expected two deployments for the app")

				// Verify deployments are ordered descending by date
				require.True(appDeployments[0].CreatedAt.After(appDeployments[1].CreatedAt), "Expected deployments to be ordered descending by date")

				// Get Deployments for Env
				envDeployments, err := stores.DeploymentStore.GetForEnv(env.Id)
				require.NoError(err, "Failed to get deployments for env")
				require.Equal(2, len(envDeployments), "Expected two deployments for the env")

				// Get Deployments for Cell
				cellDeployments, err := stores.DeploymentStore.GetForCell(cell.Id)
				require.NoError(err, "Failed to get deployments for cell")
				require.Equal(2, len(cellDeployments), "Expected two deployments for the cell")

				// Delete Deployment
				err = stores.DeploymentStore.DeleteDeployment(app.Id, env.Id, deployment.Id)
				require.NoError(err, "Failed to delete deployment")

				// Verify deployment is deleted
				_, err = stores.DeploymentStore.Get(app.Id, env.Id, deployment.Id)
				require.Error(err, "Expected error when getting deleted deployment")
			})
		})

		t.Run("Waitlist Operations", func(t *testing.T) {
			require := require.New(t)
			ctx := context.Background()

			// Test adding a new email to the waitlist
			email := fmt.Sprintf("test%d@example.com", rand.Intn(10000))
			err := stores.WaitlistStore.Add(ctx, email)
			require.NoError(err, "Failed to add email to waitlist")

			// Test adding a duplicate email to the waitlist
			err = stores.WaitlistStore.Add(ctx, email)
			require.Error(err, "Expected error when adding duplicate email to waitlist")
			require.ErrorIs(err, ErrDuplicateWaitlistEntry, "Expected ErrDuplicateWaitlistEntry error")

			// Test adding an invalid email to the waitlist
			invalidEmail := "invalid-email"
			err = stores.WaitlistStore.Add(ctx, invalidEmail)
			require.Error(err, "Expected error when adding invalid email to waitlist")
			require.Contains(err.Error(), "not a valid email", "Expected invalid email error message")
		})
	}
}
