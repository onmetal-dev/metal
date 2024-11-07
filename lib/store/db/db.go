package db

import (
	"fmt"

	"github.com/onmetal-dev/metal/lib/store"
	"go.opentelemetry.io/otel/sdk/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func open(host, user, password, dbname string, port int, sslmode string, tp *trace.TracerProvider) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", host, user, password, dbname, port)
	if sslmode != "" {
		dsn += fmt.Sprintf(" sslmode=%s", sslmode)
	}
	d, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}
	if tp != nil {
		if err := d.Use(tracing.NewPlugin(tracing.WithTracerProvider(tp), tracing.WithoutMetrics())); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func MustOpen(host, user, password, dbname string, port int, sslmode string, tp *trace.TracerProvider) *gorm.DB {
	db, err := open(host, user, password, dbname, port, sslmode, tp)
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
		&store.App{},
		&store.AppSettings{},
		&store.Env{},
		&store.AppEnvVars{},
		&store.Deployment{},
		&store.ApiToken{},
		&store.Build{},
	)
	if err != nil {
		panic(err)
	}

	return db
}
