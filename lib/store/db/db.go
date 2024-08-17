package db

import (
	"fmt"

	"github.com/onmetal-dev/metal/lib/store"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func open(host, user, password, dbname string, port int) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, user, password, dbname, port)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
}

func MustOpen(host, user, password, dbname string, port int) *gorm.DB {
	db, err := open(host, user, password, dbname, port)
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(
		&store.User{},
		&store.WaitlistedUser{},
		&store.InvitedUser{},
		&store.Team{},
		&store.TeamMember{},
		&store.TeamMemberInvite{},
		&store.PaymentMethod{},
		&store.Server{},
		&store.Cell{},
		&store.TalosCellData{},
	)
	if err != nil {
		panic(err)
	}

	return db
}
