package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/form"
	"github.com/onmetal-dev/metal/lib/store"
)

type PostApiTokenHandler struct {
	teamStore     store.TeamStore
	apiTokenStore store.ApiTokenStore
}

func NewPostApiTokenHandler(teamStore store.TeamStore, apiTokenStore store.ApiTokenStore) *PostApiTokenHandler {
	return &PostApiTokenHandler{
		teamStore:     teamStore,
		apiTokenStore: apiTokenStore,
	}
}

func (h *PostApiTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, _ := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}

	var f templates.ApiTokenFormData
	inputErrs, err := form.Decode(&f, r)
	if inputErrs.NotNil() || err != nil {
		if err := templates.ApiTokenForm(teamId, f, inputErrs, err).Render(ctx, w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if _, err := h.apiTokenStore.Create(teamId, user.Id, f.Name, f.Scope); err != nil {
		if err := templates.ApiTokenForm(teamId, f, form.FieldErrors{}, err).Render(ctx, w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/dashboard/%s/settings", teamId))
	w.WriteHeader(http.StatusOK)
}

type DeleteApiTokenHandler struct {
	teamStore     store.TeamStore
	apiTokenStore store.ApiTokenStore
}

func NewDeleteApiTokenHandler(teamStore store.TeamStore, apiTokenStore store.ApiTokenStore) *DeleteApiTokenHandler {
	return &DeleteApiTokenHandler{
		teamStore:     teamStore,
		apiTokenStore: apiTokenStore,
	}
}

func (h *DeleteApiTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, _ := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}
	apiTokenId := chi.URLParam(r, "apiTokenId")

	if err := h.apiTokenStore.Delete(apiTokenId); err != nil {
		http.Error(w, fmt.Sprintf("error deleting invite: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/dashboard/%s/settings", teamId))
	w.WriteHeader(http.StatusOK)
}
