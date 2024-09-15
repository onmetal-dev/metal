package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/cmd/app/templates"
	"github.com/onmetal-dev/metal/lib/form"
	"github.com/onmetal-dev/metal/lib/store"
)

type PostInviteHandler struct {
	userStore                      store.UserStore
	teamStore                      store.TeamStore
	loopsApiKey                    string
	loopsTxAddedToTeamNewUser      string
	loopsTxAddedToTeamExistingUser string
}

func NewPostInviteHandler(userStore store.UserStore, teamStore store.TeamStore, loopsApiKey, loopsTxAddedToTeamNewUser, loopsTxAddedToTeamExistingUser string) *PostInviteHandler {
	return &PostInviteHandler{
		userStore:                      userStore,
		teamStore:                      teamStore,
		loopsApiKey:                    loopsApiKey,
		loopsTxAddedToTeamNewUser:      loopsTxAddedToTeamNewUser,
		loopsTxAddedToTeamExistingUser: loopsTxAddedToTeamExistingUser,
	}
}

func (h *PostInviteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamId := chi.URLParam(r, "teamId")
	user := middleware.GetUser(ctx)
	team, _ := validateAndFetchTeams(ctx, h.teamStore, w, teamId, user)
	if team == nil {
		return
	}

	var f templates.InviteFormData
	inputErrs, err := form.Decode(&f, r)
	if inputErrs.NotNil() || err != nil {
		if err := templates.InviteForm(teamId, f, inputErrs, err).Render(ctx, w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if err := h.teamStore.CreateTeamInvite(f.Email, teamId); err != nil {
		if err := templates.InviteForm(teamId, f, form.FieldErrors{}, err).Render(ctx, w); err != nil {
			http.Error(w, fmt.Sprintf("error rendering template: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Send email using Loops API
	if err := h.sendInviteEmail(*team, f.Email); err != nil {
		// Log the error, but don't stop the flow
		fmt.Printf("Error sending invite email: %v\n", err)
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/dashboard/%s/settings", teamId))
	w.WriteHeader(http.StatusOK)
}

func (h *PostInviteHandler) sendInviteEmail(team store.Team, email string) error {
	// determine if this is a new or existing user, since we send different invite emails for each
	user, err := h.userStore.GetUser(email)
	if err != nil {
		return fmt.Errorf("error fetching user: %w", err)
	}
	txId := h.loopsTxAddedToTeamExistingUser
	if user == nil {
		txId = h.loopsTxAddedToTeamNewUser
	}

	client := &http.Client{}
	data := map[string]interface{}{
		"transactionalId": txId,
		"email":           email,
		"dataVariables": map[string]string{
			"email":    email,
			"teamId":   team.Id,
			"teamName": team.Name,
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", "https://app.loops.so/api/v1/transactional", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.loopsApiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

type DeleteInviteHandler struct {
	teamStore store.TeamStore
}

func NewDeleteInviteHandler(teamStore store.TeamStore) *DeleteInviteHandler {
	return &DeleteInviteHandler{
		teamStore: teamStore,
	}
}

func (h *DeleteInviteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	teamId := chi.URLParam(r, "teamId")
	email := chi.URLParam(r, "email")

	if err := h.teamStore.DeleteTeamInvite(email, teamId); err != nil {
		http.Error(w, fmt.Sprintf("error deleting invite: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/dashboard/%s/settings", teamId))
	w.WriteHeader(http.StatusOK)
}
