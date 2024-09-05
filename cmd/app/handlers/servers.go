package handlers

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/background"
	"github.com/onmetal-dev/metal/lib/background/serverfulfillment"
	"github.com/onmetal-dev/metal/lib/billing"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/samber/lo"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/billing/meter"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/price"
	"github.com/stripe/stripe-go/v79/product"
)

type GetServersNewHandler struct {
	teamStore           store.TeamStore
	serverOfferingStore store.ServerOfferingStore
}

func NewGetServersNewHandler(teamStore store.TeamStore, serverOfferingStore store.ServerOfferingStore) *GetServersNewHandler {
	return &GetServersNewHandler{
		teamStore:           teamStore,
		serverOfferingStore: serverOfferingStore,
	}
}

// validateAndFetchTeams protects a route that has a team id by validating that the user is a member of the team
// it returns non-nil if the user is a member of the team
// it takes care of responding to the http request as well
// this should probably be middleware or something... potentially cached (but that might be a headache to manage cache invalidation)
func validateAndFetchTeams(ctx context.Context, teamStore store.TeamStore, w http.ResponseWriter, teamId string, user *store.User) (*store.Team, []store.Team) {
	if !lo.ContainsBy(user.TeamMemberships, func(m store.TeamMember) bool {
		return m.TeamId == teamId
	}) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return nil, nil
	}

	team, err := teamStore.GetTeam(ctx, teamId)
	if err != nil {
		http.Error(w, "error fetching team", http.StatusInternalServerError)
		return nil, nil
	} else if team == nil {
		http.Error(w, "team doesn't exist", http.StatusNotFound)
		return nil, nil
	}

	userTeams := make([]store.Team, len(user.TeamMemberships))
	for i, m := range user.TeamMemberships {
		team, err := teamStore.GetTeam(ctx, m.TeamId)
		if err != nil || team == nil {
			http.Error(w, "error fetching team", http.StatusInternalServerError)
			return nil, nil
		}
		userTeams[i] = *team
	}

	return team, userTeams
}

func (h *GetServersNewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, userTeams := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	serverOfferings, err := h.serverOfferingStore.GetServerOfferings()
	if err != nil {
		http.Error(w, "error fetching server offerings", http.StatusInternalServerError)
		return
	}

	if err := templates.DashboardLayout(templates.DashboardState{
		User:          *user,
		UserTeams:     userTeams,
		ActiveTeam:    *team,
		ActiveTabName: templates.TabNameBuyServer,
	}, templates.CreateServer(teamId, serverOfferings)).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}

type GetServersCheckoutHandler struct {
	teamStore             store.TeamStore
	serverOfferingStore   store.ServerOfferingStore
	stripeCheckoutSession *session.Client
	stripeProduct         *product.Client
	stripePrice           *price.Client
	stripeMeter           *meter.Client
	stripePublishableKey  string
}

func NewGetServersCheckoutHandler(teamStore store.TeamStore, serverOfferingStore store.ServerOfferingStore, stripeCheckoutSession *session.Client, stripeProduct *product.Client, stripePrice *price.Client, stripeMeter *meter.Client, stripePublishableKey string) *GetServersCheckoutHandler {
	return &GetServersCheckoutHandler{
		teamStore:             teamStore,
		stripeCheckoutSession: stripeCheckoutSession,
		stripeProduct:         stripeProduct,
		stripePrice:           stripePrice,
		stripeMeter:           stripeMeter,
		stripePublishableKey:  stripePublishableKey,
		serverOfferingStore:   serverOfferingStore,
	}
}

func findOrCreateHourlyMeterForOffering(offering store.ServerOffering, locationId string, stripeMeter *meter.Client) (*stripe.BillingMeter, error) {
	// there is currently no search functionality in this api, so gotta list 'em all
	eventName := billing.UsageHourMeterEventName(offering, locationId)
	meters := stripeMeter.List(nil)
	for meters.Next() {
		meter := meters.BillingMeter()
		if meter.EventName == eventName {
			return meter, nil
		}
	}
	if err := meters.Err(); err != nil {
		return nil, fmt.Errorf("error listing meters: %w", err)
	}

	meter, err := stripeMeter.New(&stripe.BillingMeterParams{
		DisplayName: stripe.String(eventName),
		EventName:   stripe.String(eventName),
		DefaultAggregation: &stripe.BillingMeterDefaultAggregationParams{
			Formula: stripe.String(string(stripe.BillingMeterDefaultAggregationFormulaSum)),
		},
		CustomerMapping: &stripe.BillingMeterCustomerMappingParams{
			EventPayloadKey: stripe.String("stripe_customer_id"),
			Type:            stripe.String(string(stripe.BillingMeterCustomerMappingTypeByID)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating stripe meter: %w", err)
	}
	return meter, nil
}

func createStripeProductsForOffering(offering store.ServerOffering, locationId string, stripeProduct *product.Client, stripePrice *price.Client, stripeMeter *meter.Client) ([]*stripe.Product, error) {
	price, ok := lo.Find(offering.Prices, func(p store.Price) bool {
		return p.LocationId == locationId
	})
	if !ok {
		return nil, fmt.Errorf("price for location %s not found in offering %s", locationId, offering.Id)
	}

	products := []*stripe.Product{}
	metadata := map[string]string{
		"offeringId": offering.Id,
		"locationId": locationId,
	}

	if price.Setup > 0 {
		p, err := stripeProduct.New(&stripe.ProductParams{
			Name:     stripe.String(strings.ToLower(fmt.Sprintf("%s %s (%s) one-time setup fee", offering.ProviderSlug, offering.Id, locationId))),
			Metadata: metadata,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating stripe product: %w", err)
		}
		sp, err := stripePrice.New(&stripe.PriceParams{
			Product:           stripe.String(p.ID),
			UnitAmountDecimal: stripe.Float64(truncateToFourDecimalPlaces(price.Setup * 100.0)),
			Currency:          stripe.String(strings.ToLower(string(price.Currency))),
			Metadata:          metadata,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating stripe price: %w", err)
		}
		if _, err := stripeProduct.Update(p.ID, &stripe.ProductParams{
			DefaultPrice: stripe.String(sp.ID),
		}); err != nil {
			return nil, fmt.Errorf("error updating stripe product: %w", err)
		}
		pWithPrice, err := stripeProduct.Get(p.ID, &stripe.ProductParams{
			Expand: []*string{stripe.String("default_price")},
		})
		if err != nil {
			return nil, fmt.Errorf("error fetching stripe product: %w", err)
		}
		products = append(products, pWithPrice)
	}

	switch {
	case price.Hourly > 0:
		p, err := stripeProduct.New(&stripe.ProductParams{
			Name:      stripe.String(strings.ToLower(fmt.Sprintf("%s %s (%s)", offering.ProviderSlug, offering.Id, locationId))),
			UnitLabel: stripe.String("hour"),
			Metadata:  metadata,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating stripe product: %w", err)
		}
		meter, err := findOrCreateHourlyMeterForOffering(offering, locationId, stripeMeter)
		if err != nil {
			return nil, fmt.Errorf("error creating stripe meter: %w", err)
		}
		sp, err := stripePrice.New(&stripe.PriceParams{
			Product:  stripe.String(p.ID),
			Currency: stripe.String(strings.ToLower(string(price.Currency))),
			Recurring: &stripe.PriceRecurringParams{
				Interval:      stripe.String(string(stripe.PriceRecurringIntervalMonth)),
				IntervalCount: stripe.Int64(1),
				Meter:         stripe.String(meter.ID),
				UsageType:     stripe.String(string(stripe.PriceRecurringUsageTypeMetered)),
			},
			UnitAmountDecimal: stripe.Float64(truncateToFourDecimalPlaces(100.0 * price.Hourly)),
			Metadata:          metadata,
		})
		if err != nil {
			return nil, fmt.Errorf("error creating stripe price: %w", err)
		}
		if _, err := stripeProduct.Update(p.ID, &stripe.ProductParams{
			DefaultPrice: stripe.String(sp.ID),
		}); err != nil {
			return nil, fmt.Errorf("error updating stripe product: %w", err)
		}
		pWithPrice, err := stripeProduct.Get(p.ID, &stripe.ProductParams{
			Expand: []*string{stripe.String("default_price")},
		})
		if err != nil {
			return nil, fmt.Errorf("error fetching stripe product: %w", err)
		}
		products = append(products, pWithPrice)
	default:
		return nil, fmt.Errorf("unsupported price type for location %s in offering %s", locationId, offering.Id)
	}
	return products, nil
}

func (h *GetServersCheckoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, userTeams := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	offeringId := r.URL.Query().Get("offeringId")
	if offeringId == "" {
		http.Error(w, "offeringId is required", http.StatusBadRequest)
		return
	}
	locationId := r.URL.Query().Get("locationId")
	if locationId == "" {
		http.Error(w, "locationId is required", http.StatusBadRequest)
		return
	}
	offering, err := h.serverOfferingStore.GetServerOffering(offeringId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching server offering: %v", err), http.StatusInternalServerError)
		return
	}

	// find or create stripe product(s) (more than one if a setup fee)
	products := []*stripe.Product{}
	search := h.stripeProduct.Search(&stripe.ProductSearchParams{
		SearchParams: stripe.SearchParams{
			Query: fmt.Sprintf("active:'true' AND metadata['offeringId']:'%s' AND metadata['locationId']:'%s'", offeringId, locationId),
		},
		Expand: []*string{stripe.String("data.default_price")},
	})
	for search.Next() {
		products = append(products, search.Product())
	}
	if err := search.Err(); err != nil {
		http.Error(w, fmt.Sprintf("error searching for product: %v", err), http.StatusInternalServerError)
		return
	}

	if len(products) == 0 {
		fmt.Println("DEBUG creating products since none found")
		ps, err := createStripeProductsForOffering(*offering, locationId, h.stripeProduct, h.stripePrice, h.stripeMeter)
		if err != nil {
			http.Error(w, fmt.Sprintf("error creating stripe products: %v", err), http.StatusInternalServerError)
			return
		}
		products = ps
	}

	lineItems := []*stripe.CheckoutSessionLineItemParams{}
	for _, product := range products {
		price := product.DefaultPrice
		if price == nil {
			http.Error(w, fmt.Sprintf("product %s has no default price", product.ID), http.StatusInternalServerError)
			return
		}
		lineItem := &stripe.CheckoutSessionLineItemParams{
			Price: stripe.String(price.ID),
		}
		if product.DefaultPrice.Recurring == nil {
			lineItem.Quantity = stripe.Int64(1)
		}
		lineItems = append(lineItems, lineItem)
	}

	// set up embedded stripe checkout
	proto := "https"
	if strings.Contains(r.Host, "localhost") {
		proto = "http"
	}
	domain := fmt.Sprintf("%s://%s", proto, r.Host)
	returnUrl := fmt.Sprintf("%s/dashboard/%s/servers/checkout-return-url", domain, teamId)
	checkoutSession, err := h.stripeCheckoutSession.New(&stripe.CheckoutSessionParams{
		UIMode:    stripe.String("embedded"),
		ReturnURL: stripe.String(fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}", returnUrl)),
		LineItems: lineItems,
		Customer:  stripe.String(team.StripeCustomerId),
		CustomerUpdate: &stripe.CheckoutSessionCustomerUpdateParams{
			Address: stripe.String("auto"),
		},
		SavedPaymentMethodOptions: &stripe.CheckoutSessionSavedPaymentMethodOptionsParams{
			AllowRedisplayFilters: []*string{stripe.String(string(stripe.CheckoutSessionSavedPaymentMethodOptionsAllowRedisplayFilterAlways))},
		},
		Mode:         stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(true)},
		Metadata: map[string]string{
			"teamId":     teamId,
			"offeringId": offeringId,
			"locationId": locationId,
		},
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating checkout session: %v", err), http.StatusInternalServerError)
		return
	}

	// see https://docs.stripe.com/security/guide?csp=csp-js
	nonce := generateNonce()
	csp := []string{
		"default-src 'self'",
		fmt.Sprintf("script-src 'self' 'nonce-%s' https://*.js.stripe.com https://js.stripe.com https://maps.googleapis.com", nonce),
		"connect-src 'self' https://api.stripe.com https://maps.googleapis.com",
		"frame-src 'self' https://*.js.stripe.com https://js.stripe.com https://hooks.stripe.com",
		"img-src 'self' data: https:",      // Allow images from any HTTPS source and data URIs
		"style-src 'self' 'unsafe-inline'", // Allow inline styles if needed
	}
	w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))

	if err := templates.DashboardLayout(templates.DashboardState{
		User:          *user,
		UserTeams:     userTeams,
		ActiveTeam:    *team,
		ActiveTabName: templates.TabNameBuyServer,
		AdditionalScripts: []templates.ScriptTag{
			{Src: "https://js.stripe.com/v3/"},
		},
	}, templates.CreateServerCheckout(nonce, h.stripePublishableKey, checkoutSession.ClientSecret)).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}

type GetServersCheckoutReturnHandler struct {
	teamStore              store.TeamStore
	serverOfferingStore    store.ServerOfferingStore
	stripeCheckoutSession  *session.Client
	serverFulfillmentQueue *background.QueueProducer[serverfulfillment.Message]
}

func NewGetServersCheckoutReturnHandler(teamStore store.TeamStore, serverOfferingStore store.ServerOfferingStore, stripeCheckoutSession *session.Client, serverFulfillmentQueue *background.QueueProducer[serverfulfillment.Message]) *GetServersCheckoutReturnHandler {
	return &GetServersCheckoutReturnHandler{
		teamStore:              teamStore,
		serverOfferingStore:    serverOfferingStore,
		stripeCheckoutSession:  stripeCheckoutSession,
		serverFulfillmentQueue: serverFulfillmentQueue,
	}
}

func (h *GetServersCheckoutReturnHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, _ := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}

	checkoutSessionId := r.URL.Query().Get("session_id")
	if checkoutSessionId == "" {
		http.Error(w, "expected a stripe checkout session_id", http.StatusBadRequest)
		return
	}

	checkoutSession, err := h.stripeCheckoutSession.Get(checkoutSessionId, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching checkout session: %v", err), http.StatusInternalServerError)
		return
	}

	offeringId, ok := checkoutSession.Metadata["offeringId"]
	if !ok {
		http.Error(w, "offeringId not found in checkout session metadata", http.StatusBadRequest)
		return
	}
	locationId, ok := checkoutSession.Metadata["locationId"]
	if !ok {
		http.Error(w, "locationId not found in checkout session metadata", http.StatusBadRequest)
		return
	}

	if err := h.serverFulfillmentQueue.Send(context.Background(), serverfulfillment.Message{
		TeamId:                  teamId,
		UserId:                  user.Id,
		OfferingId:              offeringId,
		LocationId:              locationId,
		StripeCheckoutSessionId: checkoutSessionId,
	}); err != nil {
		http.Error(w, fmt.Sprintf("error enqueuing server fulfillment: %v", err), http.StatusInternalServerError)
		return
	}

	middleware.AddFlash(ctx, "success! hold tight while we spin up your new server")
	http.Redirect(w, r, fmt.Sprintf("/dashboard/%s", teamId), http.StatusFound)
}

// truncateToFourDecimalPlaces truncates a float64 to four decimal places so that users can comprehend prices like 0.0754 / hour
func truncateToFourDecimalPlaces(value float64) float64 {
	return math.Floor(value*10000) / 10000
}
