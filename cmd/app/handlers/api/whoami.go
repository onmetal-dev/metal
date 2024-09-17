package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/lib/store"
)

type WhoamiResponse struct {
	TokenID   string `json:"token_id"`
	TeamID    string `json:"team_id"`
	TeamName  string `json:"team_name"`
	CreatedAt string `json:"created_at"`
}

func NewWhoamiHandler(apiTokenStore store.ApiTokenStore, teamStore store.TeamStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := middleware.MustGetApiToken(r.Context())

		team, err := teamStore.GetTeam(r.Context(), token.TeamId)
		if err != nil {
			http.Error(w, "Error fetching team info", http.StatusInternalServerError)
			return
		}

		response := WhoamiResponse{
			TokenID:   token.Id,
			TeamID:    token.TeamId,
			TeamName:  team.Name,
			CreatedAt: token.CreatedAt.Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
