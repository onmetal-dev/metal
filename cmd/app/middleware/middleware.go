package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/onmetal-dev/metal/cmd/app/urls"
	"github.com/onmetal-dev/metal/lib/store"
)

type key string

var NonceKey key = "nonces"

type Nonces struct {
	Tw          string
	HtmxCssHash string
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// CSPMiddleware adds a Content Security Policy (CSP) header to the response
// to prevent XSS attacks.
func CSPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a new Nonces struct for every request when here.
		// move to outside the handler to use the same nonces in all responses
		nonceSet := Nonces{
			Tw:          generateRandomString(16),
			HtmxCssHash: "sha256-bsV5JivYxvGywDAZ22EZJKBFip65Ng9xoJVLbBg7bdo=",
		}

		// set nonces in context
		ctx := context.WithValue(r.Context(), NonceKey, nonceSet)
		cspHeader := fmt.Sprintf("default-src 'self'; script-src 'self' 'unsafe-eval' 'unsafe-inline' https://unpkg.com; style-src 'self' 'nonce-%s' '%s'; img-src 'self' data:;",
			nonceSet.Tw,
			nonceSet.HtmxCssHash,
		)

		w.Header().Set("Content-Security-Policy", cspHeader)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TextHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

// get the Nonce from the context, it is a struct called Nonces,
// so we can get the nonce we need by the key, i.e. HtmxNonce
func GetNonces(ctx context.Context) Nonces {
	nonceSet := ctx.Value(NonceKey)
	if nonceSet == nil {
		log.Fatal("error getting nonce set - is nil")
	}

	nonces, ok := nonceSet.(Nonces)

	if !ok {
		log.Fatal("error getting nonce set - not ok")
	}

	return nonces
}

func GetTwNonce(ctx context.Context) string {
	nonceSet := GetNonces(ctx)
	return nonceSet.Tw
}

type AuthMiddleware struct {
	sessionStore sessions.Store
	sessionName  string
}

func NewAuthMiddleware(sessionStore sessions.Store, sessionName string) *AuthMiddleware {
	return &AuthMiddleware{
		sessionStore: sessionStore,
		sessionName:  sessionName,
	}
}

type UserContextKey string

var UserKey UserContextKey = "user"

func (m *AuthMiddleware) AddUserToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.sessionStore.Get(r, m.sessionName)
		if err != nil {
			fmt.Println("error getting session cookie", err)
			next.ServeHTTP(w, r)
			return
		}

		user, ok := session.Values["user"].(store.User)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireLoggedInUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if user == nil {
			http.Redirect(w, r, urls.Login.Render()+"?next="+r.URL.RequestURI(), http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetUser(ctx context.Context) *store.User {
	user, ok := ctx.Value(UserKey).(store.User)
	if !ok {
		return nil
	}
	return &user
}

type FlashMiddleware struct {
	sessionStore sessions.Store
	sessionName  string
}

func NewFlashMiddleware(sessionStore sessions.Store, sessionName string) *FlashMiddleware {
	return &FlashMiddleware{
		sessionStore: sessionStore,
		sessionName:  sessionName,
	}
}

type FlashKey string

var AddFlashContextKey FlashKey = "add_flash"
var GetFlashesContextKey FlashKey = "get_flashes"

func (m *FlashMiddleware) AddFlashMethodsToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), AddFlashContextKey, func(msg string) {
			session, err := m.sessionStore.Get(r, m.sessionName)
			if err != nil {
				return
			}
			session.AddFlash(msg)
			session.Save(r, w)
		})
		ctx = context.WithValue(ctx, GetFlashesContextKey, func() []string {
			session, err := m.sessionStore.Get(r, m.sessionName)
			if err != nil {
				return []string{}
			}
			flashes := []string{}
			for _, flash := range session.Flashes() {
				str, ok := flash.(string)
				if !ok {
					continue
				}
				flashes = append(flashes, str)
			}
			session.Save(r, w)
			return flashes
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AddFlash(ctx context.Context, msg string) {
	flash, ok := ctx.Value(AddFlashContextKey).(func(msg string))
	if !ok {
		return
	}
	flash(msg)
}

func GetFlashes(ctx context.Context) []string {
	getFlashes, ok := ctx.Value(GetFlashesContextKey).(func() []string)
	if !ok {
		return []string{}
	}
	return getFlashes()
}

type apiTokenKey string

const apiTokenContextKey apiTokenKey = "api_token"

func ApiAuthMiddleware(apiTokenStore store.ApiTokenStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			bearerToken := strings.TrimPrefix(authHeader, "Bearer ")
			if bearerToken == authHeader {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			token, err := apiTokenStore.GetByToken(bearerToken)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), apiTokenContextKey, *token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func MustGetApiToken(ctx context.Context) store.ApiToken {
	token, ok := ctx.Value(apiTokenContextKey).(store.ApiToken)
	if !ok {
		panic("api token not found in context")
	}
	return token
}

func WithApiToken(ctx context.Context, token store.ApiToken) context.Context {
	return context.WithValue(ctx, apiTokenContextKey, token)
}
