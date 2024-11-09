package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/cmd/app/urls"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/customer"
	"github.com/stripe/stripe-go/v79/customersession"
	"github.com/stripe/stripe-go/v79/setupintent"
)

type GetOnboardingHandler struct {
	teamStore store.TeamStore
}

func NewGetOnboardingHandler(teamStore store.TeamStore) *GetOnboardingHandler {
	return &GetOnboardingHandler{
		teamStore: teamStore,
	}
}

func (h *GetOnboardingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUser(ctx)
	if err := templates.Layout(templates.Onboarding(*user), "metal | onboarding").Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}

type PostOnboardingHandler struct {
	teamStore store.TeamStore
}

func NewPostOnboardingHandler(teamStore store.TeamStore) *PostOnboardingHandler {
	return &PostOnboardingHandler{
		teamStore: teamStore,
	}
}

func (h *PostOnboardingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := middleware.GetUser(ctx)
	teamName := r.FormValue("team-name")
	if teamName == "" || len(teamName) < 5 {
		http.Error(w, "team name is required, and should be at least 5 characters long", http.StatusBadRequest)
		return
	}
	team, err := h.teamStore.CreateTeam(teamName, "")
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating team: %v", err), http.StatusInternalServerError)
		return
	}
	if err := h.teamStore.AddUserToTeam(user.Id, team.Id); err != nil {
		http.Error(w, fmt.Sprintf("error adding user to team: %v", err), http.StatusInternalServerError)
		return
	}
	if err := h.teamStore.CreateStripeCustomer(ctx, team.Id, user.Email); err != nil {
		http.Error(w, fmt.Sprintf("error creating stripe customer: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", urls.OnboardingPayment{TeamId: team.Id}.Render())
	w.WriteHeader(http.StatusOK)
}

type GetOnboardingPaymentHandler struct {
	teamStore             store.TeamStore
	stripeCustomerSession *customersession.Client
}

func NewGetOnboardingPaymentHandler(teamStore store.TeamStore, stripeCustomerSession *customersession.Client) *GetOnboardingPaymentHandler {
	return &GetOnboardingPaymentHandler{
		teamStore:             teamStore,
		stripeCustomerSession: stripeCustomerSession,
	}
}

func generateNonce() string {
	nonce := make([]byte, 16)
	rand.Read(nonce)
	return base64.StdEncoding.EncodeToString(nonce)
}

func (h *GetOnboardingPaymentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	team, err := h.teamStore.GetTeam(ctx, teamId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting team: %v", err), http.StatusInternalServerError)
		return
	}
	// see https://docs.stripe.com/api/customer_sessions/create#create_customer_session-components-payment_element-features
	csParams := &stripe.CustomerSessionParams{
		Customer:   stripe.String(team.StripeCustomerId),
		Components: &stripe.CustomerSessionComponentsParams{},
	}
	csParams.AddExtra("components[payment_element][enabled]", "true")
	csParams.AddExtra(
		"components[payment_element][features][payment_method_save]",
		"enabled",
	)
	csParams.AddExtra(
		"components[payment_element][features][payment_method_save_usage]",
		"off_session",
	)
	customerSession, err := h.stripeCustomerSession.New(csParams)
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating stripe customer session: %v", err), http.StatusInternalServerError)
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

	if err := templates.Layout(templates.OnboardingPayment(nonce, team.Id, customerSession.ClientSecret), "metal | onboarding", templates.ScriptTag{Src: "https://js.stripe.com/v3/"}).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}

type PostOnboardingPaymentHandler struct {
	teamStore         store.TeamStore
	stripeSetupIntent *setupintent.Client
}

func NewPostOnboardingPaymentHandler(teamStore store.TeamStore, stripeSetupIntent *setupintent.Client) *PostOnboardingPaymentHandler {
	return &PostOnboardingPaymentHandler{
		teamStore:         teamStore,
		stripeSetupIntent: stripeSetupIntent,
	}
}

func (h *PostOnboardingPaymentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	team, err := h.teamStore.GetTeam(ctx, teamId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting team: %v", err), http.StatusInternalServerError)
		return
	}
	// see https://docs.stripe.com/payments/accept-a-payment-deferred?platform=web&type=setup&client=html#create-intent
	params := &stripe.SetupIntentParams{
		// To allow saving and retrieving payment methods, provide the Customer ID.
		Customer: stripe.String(team.StripeCustomerId),
		AutomaticPaymentMethods: &stripe.SetupIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}
	intent, _ := h.stripeSetupIntent.New(params)
	data := map[string]string{
		"client_secret": intent.ClientSecret,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

type GetOnboardingPaymentConfirmHandler struct {
	teamStore         store.TeamStore
	stripeSetupIntent *setupintent.Client
	stripeCustomer    *customer.Client
}

func NewGetOnboardingPaymentConfirmHandler(teamStore store.TeamStore, stripeSetupIntent *setupintent.Client, stripeCustomer *customer.Client) *GetOnboardingPaymentConfirmHandler {
	return &GetOnboardingPaymentConfirmHandler{
		teamStore:         teamStore,
		stripeSetupIntent: stripeSetupIntent,
		stripeCustomer:    stripeCustomer,
	}
}

func (h *GetOnboardingPaymentConfirmHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	redirectStatus := r.URL.Query().Get("redirect_status")
	if redirectStatus != "succeeded" {
		http.Error(w, "error redirect status", http.StatusInternalServerError)
		return
	}
	teamId := chi.URLParam(r, "teamId")
	team, err := h.teamStore.GetTeam(ctx, teamId)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting team: %v", err), http.StatusInternalServerError)
		return
	}
	setupIntent := r.URL.Query().Get("setup_intent")
	si, err := h.stripeSetupIntent.Get(setupIntent, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting setup intent: %v", err), http.StatusInternalServerError)
		return
	} else if si.PaymentMethod == nil {
		http.Error(w, "error payment method is nil", http.StatusInternalServerError)
		return
	}
	pm, err := h.stripeCustomer.RetrievePaymentMethod(si.PaymentMethod.ID, &stripe.CustomerRetrievePaymentMethodParams{Customer: stripe.String(team.StripeCustomerId)})
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting payment method: %v", err), http.StatusInternalServerError)
		return
	}
	if err := h.teamStore.AddPaymentMethod(ctx, teamId, store.PaymentMethod{
		StripePaymentMethodId: pm.ID,
		Default:               true, // this is the first pm they've set up, so make it default
		Type:                  string(pm.Card.Brand),
		Last4:                 pm.Card.Last4,
		ExpirationMonth:       int(pm.Card.ExpMonth),
		ExpirationYear:        int(pm.Card.ExpYear),
	}); err != nil {
		http.Error(w, fmt.Sprintf("error adding payment method: %v", err), http.StatusInternalServerError)
		return
	}

	nonce := generateNonce()
	confettiUrl := "https://cdn.jsdelivr.net/npm/canvas-confetti@1.9.3/dist/confetti.browser.min.js"
	csp := fmt.Sprintf("script-src 'nonce-%s' %s 'self'; style-src 'self' 'unsafe-inline'; worker-src 'self' blob:", nonce, confettiUrl)
	w.Header().Set("Content-Security-Policy", csp)
	if err := templates.Layout(templates.OnboardingSuccess(nonce, teamId), "metal | onboarding", templates.ScriptTag{Src: confettiUrl}).Render(ctx, w); err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
	}
}
