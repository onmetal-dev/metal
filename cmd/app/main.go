package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/dnsprovider"
	"github.com/onmetal-dev/metal/lib/logger"
	"github.com/onmetal-dev/metal/lib/serverprovider"
	"github.com/onmetal-dev/metal/lib/store"
	database "github.com/onmetal-dev/metal/lib/store/db"
	"github.com/onmetal-dev/metal/lib/store/dbstore"
	"github.com/onmetal-dev/metal/lib/talosprovider"
	slogformatter "github.com/samber/slog-formatter"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/billing/meter"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/customer"
	"github.com/stripe/stripe-go/v79/customersession"
	"github.com/stripe/stripe-go/v79/price"
	"github.com/stripe/stripe-go/v79/product"
	"github.com/stripe/stripe-go/v79/setupintent"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

/*
* Set to production at build time
* used to determine what assets to load
 */
var Environment = "local"

func init() {
	os.Setenv("env", Environment)
}

func initTracerProvider() (*sdktrace.TracerProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resource, err := resource.New(ctx, resource.WithAttributes(
		attribute.String("service.name", "metal"),
		attribute.String("service.namespace", "default"),
	))
	if err != nil {
		return nil, err
	}

	exporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint("localhost:4317"), // Default OTLP gRPC port
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
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
	ctx := context.Background()
	tp, _ := initTracerProvider()
	defer tp.Shutdown(ctx)

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

	db := database.MustOpen(c.DatabaseHost, c.DatabaseUser, c.DatabasePassword, c.DatabaseName, c.DatabasePort, c.DatabaseSslMode)
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

	// api clients
	hrobotClient := hrobot.NewClient(hrobot.WithToken(fmt.Sprintf("%s:%s", c.HetznerRobotUsername, c.HetznerRobotPassword)))

	// serverproviders
	serverProviderHetzner, err := serverprovider.NewHetzner(
		serverprovider.WithHrobotClient(hrobotClient),
		serverprovider.WithAuthorizedKeyFingerprint(c.SshKeyFingerprint),
	)
	if err != nil {
		slogger.Error("Failed to create server provider", slog.Any("err", err))
		os.Exit(1)
	}

	// talosproviders
	talosProviderHetzner, err := talosprovider.NewHetznerProvider(
		talosprovider.WithClient(hrobotClient),
		talosprovider.WithLogger(slogger),
	)
	if err != nil {
		slogger.Error("Failed to create talos provider", slog.Any("err", err))
		os.Exit(1)
	}

	// dnsprovider
	cfApi, err := cloudflare.NewWithAPIToken(c.CloudflareApiToken)
	if err != nil {
		slogger.Error("Failed to create cloudflare api", slog.Any("err", err))
		os.Exit(1)
	}
	cfDnsProvider, err := dnsprovider.NewCloudflareDNSProvider(dnsprovider.WithApi(cfApi), dnsprovider.WithZoneId(c.CloudflareOnmetalDotRunZoneId))
	if err != nil {
		slogger.Error("Failed to create cloudflare dns provider", slog.Any("err", err))
		os.Exit(1)
	}

	// cellprovider
	talosCellProvider, err := cellprovider.NewTalosClusterCellProvider(
		cellprovider.WithDnsProvider(cfDnsProvider),
		cellprovider.WithCellStore(cellStore),
		cellprovider.WithServerStore(serverStore),
		cellprovider.WithTmpDirRoot(c.TmpDirRoot),
		cellprovider.WithLogger(slog.Default()),
	)
	if err != nil {
		slogger.Error("Failed to create talos cell provider", slog.Any("err", err))
		os.Exit(1)
	}
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
	queueName := "fulfillment"
	producer := background.NewQueueProducer[background.ServerFulfillment](ctx, queueName, connString)
	serverFulfillmentHandler, err := background.NewServerFulfillmentHandler(
		background.WithLogger(slogger),
		background.WithQueueProducer(producer),
		background.WithTeamStore(teamStore),
		background.WithUserStore(userStore),
		background.WithServerStore(serverStore),
		background.WithServerOfferingStore(serverOfferingStore),
		background.WithCellStore(cellStore),
		background.WithStripeCheckoutSession(stripeCheckoutSession),
		background.WithServerProviderHetzner(serverProviderHetzner),
		background.WithTalosProviderHetzner(talosProviderHetzner),
		background.WithTalosCellProvider(talosCellProvider),
		background.WithSshKeyBase64(c.SshKeyBase64),
		background.WithSshKeyPassword(c.SshKeyPassword),
		background.WithSshKeyFingerprint(c.SshKeyFingerprint),
	)
	if err != nil {
		slogger.Error("Failed to create server fulfillment handler", slog.Any("err", err))
		os.Exit(1)
	}
	consumer := background.NewQueueConsumer[background.ServerFulfillment](ctx, queueName, connString, 180, serverFulfillmentHandler.Handle, slogger)
	go consumer.Start(ctx)
	defer consumer.Stop()

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
			r.Get("/dashboard/{teamId}", handlers.NewDashboardHandler(userStore, teamStore, serverStore, cellStore, cellProviderForType).ServeHTTP)
			r.Get("/dashboard/{teamId}/servers/new", handlers.NewGetServersNewHandler(teamStore, serverOfferingStore).ServeHTTP)
			r.Get("/dashboard/{teamId}/servers/checkout", handlers.NewGetServersCheckoutHandler(teamStore, serverOfferingStore, stripeCheckoutSession, stripeProduct, stripePrice, stripeMeter, c.StripePublishableKey).ServeHTTP)
			r.Get("/dashboard/{teamId}/servers/checkout-return-url", handlers.NewGetServersCheckoutReturnHandler(teamStore, serverOfferingStore, stripeCheckoutSession, producer).ServeHTTP)
		})
	})

	killSig := make(chan os.Signal, 1)

	signal.Notify(killSig, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", c.Port),
		Handler: r,
	}

	go func() {
		err := srv.ListenAndServe()

		if errors.Is(err, http.ErrServerClosed) {
			slogger.Info("Server shutdown complete")
		} else if err != nil {
			slogger.Error("Server error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	slogger.Info("Server started", slog.String("port", c.Port), slog.String("env", Environment))
	<-killSig

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
