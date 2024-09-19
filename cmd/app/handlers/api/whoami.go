package api

import (
	"context"

	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/lib/oapi"
)

func (a api) WhoAmI(ctx context.Context, request oapi.WhoAmIRequestObject) (oapi.WhoAmIResponseObject, error) {
	token := middleware.MustGetApiToken(ctx)

	team, err := a.teamStore.GetTeam(ctx, token.TeamId)
	if err != nil {
		return oapi.WhoAmI500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
	}

	return oapi.WhoAmI200JSONResponse{
		TokenId:   token.Id,
		TeamId:    token.TeamId,
		TeamName:  team.Name,
		CreatedAt: token.CreatedAt,
	}, nil
}
