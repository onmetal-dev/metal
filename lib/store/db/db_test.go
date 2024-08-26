package db

import (
	"testing"

	"github.com/onmetal-dev/metal/cmd/app/hash/passwordhash"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/onmetal-dev/metal/lib/store/dbstore"
)

func TestDB(t *testing.T) {
	host := "localhost"
	user := "postgres"
	password := "postgres"
	dbname := "metal_test"
	port := 5433
	db := MustOpen(host, user, password, dbname, port, "disable")

	userStore := dbstore.NewUserStore(dbstore.NewUserStoreParams{
		DB:           db,
		PasswordHash: passwordhash.NewHPasswordHash(),
	})
	teamStore := dbstore.NewTeamStore(dbstore.NewTeamStoreParams{
		DB: db,
	})
	serverStore := dbstore.NewServerStore(dbstore.NewServerStoreParams{
		DB: db,
	})
	cellStore := dbstore.NewCellStore(dbstore.NewCellStoreParams{
		DB: db,
	})
	appStore := dbstore.NewAppStore(dbstore.NewAppStoreParams{
		DB: db,
	})
	deploymentStore := dbstore.NewDeploymentStore(dbstore.NewDeploymentStoreParams{
		DB: db,
	})

	testSuite := store.NewStoreTestSuite(store.TestStoresConfig{
		UserStore:       userStore,
		TeamStore:       teamStore,
		ServerStore:     serverStore,
		CellStore:       cellStore,
		AppStore:        appStore,
		DeploymentStore: deploymentStore,
	})
	testSuite(t)
}
