package api

import (
	"context"

	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/lib/oapi"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/samber/lo"
	"go.jetify.com/typeid"
)

func appFromStore(app store.App) oapi.App {
	return oapi.App{
		Id:        app.Id,
		Name:      app.Name,
		CreatedAt: app.CreatedAt,
		UpdatedAt: app.UpdatedAt,
		CreatorId: app.UserId,
		TeamId:    app.TeamId,
	}
}

func appsFromStore(apps []store.App) []oapi.App {
	return lo.Map(apps, func(app store.App, _ int) oapi.App {
		return appFromStore(app)
	})
}

func (a api) GetApps(ctx context.Context, request oapi.GetAppsRequestObject) (oapi.GetAppsResponseObject, error) {
	token := middleware.MustGetApiToken(ctx)

	apps, err := a.appStore.GetForTeam(ctx, token.TeamId)
	if err != nil {
		return oapi.GetApps500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
	}

	return oapi.GetApps200JSONResponse(appsFromStore(apps)), nil
}

func (a api) DeleteApp(ctx context.Context, request oapi.DeleteAppRequestObject) (oapi.DeleteAppResponseObject, error) {
	token := middleware.MustGetApiToken(ctx)

	app, err := a.appStore.Get(ctx, request.AppId)
	if err != nil {
		if err == store.ErrAppNotFound {
			return oapi.DeleteApp404JSONResponse{NotFoundJSONResponse: oapi.NotFoundJSONResponse{Error: "not found"}}, nil
		}
		return oapi.DeleteApp500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
	} else if app.TeamId != token.TeamId {
		return oapi.DeleteApp404JSONResponse{NotFoundJSONResponse: oapi.NotFoundJSONResponse{Error: "not found"}}, nil
	}

	if err := a.appStore.Delete(ctx, request.AppId); err != nil {
		return oapi.DeleteApp500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
	}

	return oapi.DeleteApp204Response{}, nil
}

func (a api) GetApp(ctx context.Context, request oapi.GetAppRequestObject) (oapi.GetAppResponseObject, error) {
	token := middleware.MustGetApiToken(ctx)

	app, err := a.appStore.Get(ctx, request.AppId)
	if err != nil {
		if err == store.ErrAppNotFound {
			return oapi.GetApp404JSONResponse{NotFoundJSONResponse: oapi.NotFoundJSONResponse{Error: "not found"}}, nil
		}
		return oapi.GetApp500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
	} else if app.TeamId != token.TeamId {
		return oapi.GetApp404JSONResponse{NotFoundJSONResponse: oapi.NotFoundJSONResponse{Error: "not found"}}, nil
	}

	return oapi.GetApp200JSONResponse(appFromStore(app)), nil
}

func (a api) CreateApp(ctx context.Context, request oapi.CreateAppRequestObject) (oapi.CreateAppResponseObject, error) {
	token := middleware.MustGetApiToken(ctx)

	tid, err := typeid.FromString(request.AppId)
	if err != nil {
		return oapi.CreateApp400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: err.Error()}}, nil
	} else if tid.Prefix() != "app" {
		return oapi.CreateApp400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: "invalid app id"}}, nil
	}

	app, err := a.appStore.Create(store.CreateAppOptions{
		TeamId: token.TeamId,
		Name:   request.Body.Name,
		UserId: token.CreatorId,
	})
	if err != nil {
		return oapi.CreateApp500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
	}

	return oapi.CreateApp201JSONResponse{
		Id:        app.Id,
		Name:      app.Name,
		CreatedAt: app.CreatedAt,
		UpdatedAt: app.UpdatedAt,
		CreatorId: app.UserId,
		TeamId:    app.TeamId,
	}, nil
}
