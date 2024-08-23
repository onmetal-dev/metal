package db

import (
	"fmt"

	"github.com/onmetal-dev/metal/lib/store"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func open(host, user, password, dbname string, port int, sslmode string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", host, user, password, dbname, port)
	if sslmode != "" {
		dsn += fmt.Sprintf(" sslmode=%s", sslmode)
	}
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
}

func MustOpen(host, user, password, dbname string, port int, sslmode string) *gorm.DB {
	db, err := open(host, user, password, dbname, port, sslmode)
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
		&store.ServerBillingStripeUsageBasedHourly{},
		&store.Cell{},
		&store.TalosCellData{},
	)
	if err != nil {
		panic(err)
	}

	return db
}
