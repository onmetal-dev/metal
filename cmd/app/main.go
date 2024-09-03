package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/floshodan/hrobot-go/hrobot"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/gorilla/sessions"
	"github.com/onmetal-dev/metal/cmd/app/config"
	"github.com/onmetal-dev/metal/cmd/app/handlers"
	"github.com/onmetal-dev/metal/cmd/app/hash/passwordhash"
	m "github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/lib/background"
	"github.com/onmetal-dev/metal/lib/background/deployment"
	"github.com/onmetal-dev/metal/lib/background/serverbillinghourly"
	"github.com/onmetal-dev/metal/lib/background/serverfulfillment"
	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/dnsprovider"
	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/serverprovider"
	"github.com/onmetal-dev/metal/lib/store"
	database "github.com/onmetal-dev/metal/lib/store/db"
	"github.com/onmetal-dev/metal/lib/store/dbstore"
	"github.com/onmetal-dev/metal/lib/talosprovider"
	"github.com/riandyrn/otelchi"
	slogformatter "github.com/samber/slog-formatter"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/billing/meter"
	"github.com/stripe/stripe-go/v79/billing/meterevent"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/customer"
	"github.com/stripe/stripe-go/v79/customersession"
	"github.com/stripe/stripe-go/v79/price"
	"github.com/stripe/stripe-go/v79/product"
	"github.com/stripe/stripe-go/v79/setupintent"
)

/*
* Set to production at build time
* used to determine what assets to load
 */
var Environment = "local"

func init() {
	os.Setenv("env", Environment)
}

func mustCreate[T any](slogger *slog.Logger, f func() (T, error)) T {
	v, err := f()
	if err != nil {
		slogger.Error("Error", slog.Any("err", err))
		os.Exit(1)
	}
	return v
}

func main() {
	slogger := slog.New(
		slogformatter.NewFormatterHandler(
			slogformatter.TimezoneConverter(time.UTC),
			slogformatter.TimeFormatter(time.RFC3339, nil),
		)(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}),
		),
	)
	c := config.MustLoadConfig()

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, tracerProvider, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// stripe clients
	stripeBackend := stripe.NewBackends(http.DefaultClient).API
	stripeCustomer := &customer.Client{
		B:   stripeBackend,
		Key: c.StripeSecretKey,
	}
	stripeCustomerSession := &customersession.Client{
		B:   stripeBackend,
		Key: c.StripeSecretKey,
	}
	stripeSetupIntent := &setupintent.Client{
		B:   stripeBackend,
		Key: c.StripeSecretKey,
	}
	stripeCheckoutSession := &session.Client{
		B:   stripeBackend,
		Key: c.StripeSecretKey,
	}
	stripeProduct := &product.Client{
		B:   stripeBackend,
		Key: c.StripeSecretKey,
	}
	stripePrice := &price.Client{
		B:   stripeBackend,
		Key: c.StripeSecretKey,
	}
	stripeMeter := &meter.Client{
		B:   stripeBackend,
		Key: c.StripeSecretKey,
	}
	stripeMeterEvent := &meterevent.Client{
		B:   stripeBackend,
		Key: c.StripeSecretKey,
	}

	db := database.MustOpen(c.DatabaseHost, c.DatabaseUser, c.DatabasePassword, c.DatabaseName, c.DatabasePort, c.DatabaseSslMode, tracerProvider)
	passwordhash := passwordhash.NewHPasswordHash()

	waitlistStore := dbstore.NewWaitlistStore(
		dbstore.NewWaitlistStoreParams{
			LoopsWaitlistFormUrl: c.LoopsWaitlistFormUrl,
			DB:                   db,
		},
	)
	inviteStore := dbstore.NewInviteStore(
		dbstore.NewInviteStoreParams{
			DB: db,
		},
	)
	userStore := dbstore.NewUserStore(
		dbstore.NewUserStoreParams{
			DB:           db,
			PasswordHash: passwordhash,
		},
	)

	teamStore := dbstore.NewTeamStore(
		dbstore.NewTeamStoreParams{
			DB:             db,
			StripeCustomer: stripeCustomer,
		},
	)

	serverStore := dbstore.NewServerStore(
		dbstore.NewServerStoreParams{
			DB: db,
		},
	)
	serverOfferingStore := dbstore.NewServerOfferingStore(
		dbstore.NewServerOfferingStoreParams{
			//DB: db,
		},
	)
	cellStore := dbstore.NewCellStore(
		dbstore.NewCellStoreParams{
			DB: db,
		},
	)

	appStore := dbstore.NewAppStore(
		dbstore.NewAppStoreParams{
			DB: db,
		},
	)

	deploymentStore := mustCreate(slogger, func() (*dbstore.DeploymentStore, error) {
		return dbstore.NewDeploymentStore(
			dbstore.NewDeploymentStoreParams{
				DB:          db,
				GetTeamKeys: teamStore.GetTeamKeys,
			},
		)
	})

	// api clients
	hrobotClient := hrobot.NewClient(hrobot.WithToken(fmt.Sprintf("%s:%s", c.HetznerRobotUsername, c.HetznerRobotPassword)))

	// serverproviders
	serverProviderHetzner := mustCreate(slogger, func() (serverprovider.ServerProvider, error) {
		return serverprovider.NewHetzner(
			serverprovider.WithHrobotClient(hrobotClient),
			serverprovider.WithAuthorizedKeyFingerprint(c.SshKeyFingerprint),
		)
	})

	// talosproviders
	talosProviderHetzner := mustCreate(slogger, func() (*talosprovider.HetznerProvider, error) {
		return talosprovider.NewHetznerProvider(
			talosprovider.WithClient(hrobotClient),
			talosprovider.WithLogger(slogger),
		)
	})

	// dnsprovider
	cfApi := mustCreate(slogger, func() (*cloudflare.API, error) {
		return cloudflare.NewWithAPIToken(c.CloudflareApiToken)
	})
	cfDnsProvider := mustCreate(slogger, func() (dnsprovider.DNSProvider, error) {
		return dnsprovider.NewCloudflareDNSProvider(dnsprovider.WithApi(cfApi), dnsprovider.WithZoneId(c.CloudflareOnmetalDotRunZoneId))
	})

	// cellprovider
	talosCellProvider := mustCreate(slogger, func() (*cellprovider.TalosClusterCellProvider, error) {
		return cellprovider.NewTalosClusterCellProvider(
			cellprovider.WithDnsProvider(cfDnsProvider),
			cellprovider.WithCellStore(cellStore),
			cellprovider.WithServerStore(serverStore),
			cellprovider.WithTmpDirRoot(c.TmpDirRoot),
			cellprovider.WithLogger(slog.Default()),
		)
	})
	cellProviderForType := func(cellType store.CellType) cellprovider.CellProvider {
		switch cellType {
		case store.CellTypeTalos:
			return talosCellProvider
		default:
			return nil
		}
	}

	// background workers
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.DatabaseUser, c.DatabasePassword, c.DatabaseHost, c.DatabasePort, c.DatabaseName)
	if c.DatabaseSslMode != "" {
		connString += fmt.Sprintf("?sslmode=%s", c.DatabaseSslMode)
	}
	queueNameBilling := "server_billing_hourly"
	producerBilling := background.NewQueueProducer[serverbillinghourly.Message](ctx, queueNameBilling, connString)
	serverBillingHourlyHandler := mustCreate(slogger, func() (*serverbillinghourly.MessageHandler, error) {
		return serverbillinghourly.NewMessageHandler(
			serverbillinghourly.WithLogger(slogger),
			serverbillinghourly.WithQueueProducer(producerBilling),
			serverbillinghourly.WithTeamStore(teamStore),
			serverbillinghourly.WithServerStore(serverStore),
			serverbillinghourly.WithServerOfferingStore(serverOfferingStore),
			serverbillinghourly.WithStripeMeterEvent(stripeMeterEvent),
		)
	})
	{
		consumer := background.NewQueueConsumer[serverbillinghourly.Message](ctx, queueNameBilling, connString, 30, serverBillingHourlyHandler.Handle, slogger)
		go consumer.Start(ctx)
		defer consumer.Stop()
	}

	queueNameFulfillment := "fulfillment"
	producerFulfillment := background.NewQueueProducer[serverfulfillment.Message](ctx, queueNameFulfillment, connString)
	serverFulfillmentHandler := mustCreate(slogger, func() (*serverfulfillment.MessageHandler, error) {
		return serverfulfillment.NewMessageHandler(
			serverfulfillment.WithLogger(slogger),
			serverfulfillment.WithQueueProducer(producerFulfillment),
			serverfulfillment.WithServerBillingHourlyProducer(producerBilling),
			serverfulfillment.WithTeamStore(teamStore),
			serverfulfillment.WithUserStore(userStore),
			serverfulfillment.WithServerStore(serverStore),
			serverfulfillment.WithServerOfferingStore(serverOfferingStore),
			serverfulfillment.WithCellStore(cellStore),
			serverfulfillment.WithStripeCheckoutSession(stripeCheckoutSession),
			serverfulfillment.WithServerProviderHetzner(serverProviderHetzner),
			serverfulfillment.WithTalosProviderHetzner(talosProviderHetzner),
			serverfulfillment.WithTalosCellProvider(talosCellProvider),
			serverfulfillment.WithSshKeyBase64(c.SshKeyBase64),
			serverfulfillment.WithSshKeyPassword(c.SshKeyPassword),
			serverfulfillment.WithSshKeyFingerprint(c.SshKeyFingerprint),
		)
	})
	{
		consumer := background.NewQueueConsumer[serverfulfillment.Message](ctx, queueNameFulfillment, connString, 180, serverFulfillmentHandler.Handle, slogger)
		go consumer.Start(ctx)
		defer consumer.Stop()
	}

	queueNameDeployment := "deployment"
	producerDeployment := background.NewQueueProducer[deployment.Message](ctx, queueNameDeployment, connString)
	deploymentHandler := mustCreate(slogger, func() (*deployment.MessageHandler, error) {
		return deployment.NewMessageHandler(
			deployment.WithLogger(slogger),
			deployment.WithQueueProducer(producerDeployment),
			deployment.WithDeploymentStore(deploymentStore),
			deployment.WithCellProviderForType(cellProviderForType),
			deployment.WithCellStore(cellStore),
		)
	})
	{
		consumer := background.NewQueueConsumer[deployment.Message](ctx, queueNameDeployment, connString, 60, deploymentHandler.Handle, slogger)
		go consumer.Start(ctx)
		defer consumer.Stop()
	}
	// http router
	r := chi.NewRouter()
	r.Use(httprate.LimitByIP(100, time.Minute))
	fileServer := http.FileServer(http.Dir("./cmd/app/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	sessionStore := sessions.NewCookieStore([]byte(c.SessionKey))
	authMiddleware := m.NewAuthMiddleware(sessionStore, c.SessionName)
	flashMiddleware := m.NewFlashMiddleware(sessionStore, c.SessionName)

	r.Group(func(r chi.Router) {
		r.Use(
			middleware.Logger,
			m.TextHTMLMiddleware,
			m.CSPMiddleware,
			authMiddleware.AddUserToContext,
			flashMiddleware.AddFlashMethodsToContext,
			// inject the logger into each request's context
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ctx := logger.AddToContext(r.Context(), slogger)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			},
			otelchi.Middleware("metal", otelchi.WithChiRoutes(r), otelchi.WithTracerProvider(tracerProvider)),
		)

		r.NotFound(handlers.NewNotFoundHandler().ServeHTTP)

		r.Get("/", handlers.NewHomeHandler().ServeHTTP)

		r.Get("/about", handlers.NewAboutHandler().ServeHTTP)

		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		r.Post("/waitlist", handlers.NewPostWaitlistHandler(handlers.PostWaitlistHandlerParams{
			WaitlistStore: waitlistStore,
		}).ServeHTTP)
		r.Get("/signup", handlers.NewGetSignUpHandler(inviteStore).ServeHTTP)
		r.Post("/signup", handlers.NewPostSignUpHandler(userStore, inviteStore, teamStore).ServeHTTP)
		r.Get("/login", handlers.NewGetLoginHandler().ServeHTTP)
		r.Post("/login", handlers.NewPostLoginHandler(userStore, teamStore, passwordhash, sessionStore, c.SessionName).ServeHTTP)
		r.Post("/logout", handlers.NewPostLogoutHandler(handlers.PostLogoutHandlerParams{
			SessionStore: sessionStore,
			SessionName:  c.SessionName,
		}).ServeHTTP)

		// logged in routes below
		r.Group(func(r chi.Router) {
			r.Use(m.RequireLoggedInUser)
			r.Get("/onboarding", handlers.NewGetOnboardingHandler(teamStore).ServeHTTP)
			r.Post("/onboarding", handlers.NewPostOnboardingHandler(teamStore).ServeHTTP)
			r.Get("/onboarding/{teamId}/payment", handlers.NewGetOnboardingPaymentHandler(teamStore, stripeCustomerSession).ServeHTTP)
			r.Post("/onboarding/{teamId}/payment", handlers.NewPostOnboardingPaymentHandler(teamStore, stripeSetupIntent).ServeHTTP)
			r.Get("/onboarding/{teamId}/payment/confirm", handlers.NewGetOnboardingPaymentConfirmHandler(teamStore, stripeSetupIntent, stripeCustomer).ServeHTTP)
			r.Get("/dashboard/{teamId}", handlers.NewDashboardHandler(userStore, teamStore, serverStore, cellStore, deploymentStore, appStore, cellProviderForType).ServeHTTP)
			r.Get("/dashboard/{teamId}/servers/new", handlers.NewGetServersNewHandler(teamStore, serverOfferingStore).ServeHTTP)
			r.Get("/dashboard/{teamId}/servers/checkout", handlers.NewGetServersCheckoutHandler(teamStore, serverOfferingStore, stripeCheckoutSession, stripeProduct, stripePrice, stripeMeter, c.StripePublishableKey).ServeHTTP)
			r.Get("/dashboard/{teamId}/servers/checkout-return-url", handlers.NewGetServersCheckoutReturnHandler(teamStore, serverOfferingStore, stripeCheckoutSession, producerFulfillment).ServeHTTP)
			r.Get("/dashboard/{teamId}/apps/new", handlers.NewGetAppsNewHandler(userStore, teamStore, serverStore, cellStore).ServeHTTP)
			r.Post("/dashboard/{teamId}/apps/new", handlers.NewPostAppsNewHandler(userStore, teamStore, serverStore, cellStore, appStore, deploymentStore, producerDeployment).ServeHTTP)
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", c.Port),
		Handler: r,
	}

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()
	slogger.Info("Server started", slog.String("port", c.Port), slog.String("env", Environment))

	// Wait for interruption.
	select {
	case err = <-srvErr:
		if errors.Is(err, http.ErrServerClosed) {
			slogger.Info("Server shutdown complete")
		} else if err != nil {
			slogger.Error("Server error", slog.Any("err", err))
		}
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	slogger.Info("Shutting down server")

	// Create a context with a timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		slogger.Error("Server shutdown failed", slog.Any("err", err))
		os.Exit(1)
	}
	slogger.Info("Server shutdown complete")
}
