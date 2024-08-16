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
	db := MustOpen(host, user, password, dbname, port)

	userStore := dbstore.NewUserStore(dbstore.NewUserStoreParams{
		DB:           db,
		PasswordHash: passwordhash.NewHPasswordHash(),
	})
	teamStore := dbstore.NewTeamStore(dbstore.NewTeamStoreParams{
		DB: db,
	})

	testSuite := store.NewStoreTestSuite(store.TestStoresConfig{
		UserStore: userStore,
		TeamStore: teamStore,
	})

	testSuite(t)
}
