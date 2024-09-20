package api

import (
	"context"

	"github.com/onmetal-dev/metal/cmd/app/middleware"
	"github.com/onmetal-dev/metal/lib/oapi"
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/samber/lo"
	"go.jetify.com/typeid"
)

func envFromStore(env store.Env) oapi.Env {
	return oapi.Env{
		Id:        env.Id,
		Name:      env.Name,
		CreatedAt: env.CreatedAt,
		UpdatedAt: env.UpdatedAt,
	}
}

func envsFromStore(envs []store.Env) []oapi.Env {
	return lo.Map(envs, func(env store.Env, _ int) oapi.Env {
		return envFromStore(env)
	})
}

func (a api) GetEnvs(ctx context.Context, request oapi.GetEnvsRequestObject) (oapi.GetEnvsResponseObject, error) {
	token := middleware.MustGetApiToken(ctx)
	envs, err := a.deploymentStore.GetEnvsForTeam(token.TeamId)
	if err != nil {
		return oapi.GetEnvs500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
	}
	return oapi.GetEnvs200JSONResponse(envsFromStore(envs)), nil
}

func (a api) DeleteEnv(ctx context.Context, request oapi.DeleteEnvRequestObject) (oapi.DeleteEnvResponseObject, error) {
	token := middleware.MustGetApiToken(ctx)

	env, err := a.deploymentStore.GetEnv(request.EnvId)
	if err != nil {
		if err == store.ErrEnvNotFound {
			return oapi.DeleteEnv404JSONResponse{NotFoundJSONResponse: oapi.NotFoundJSONResponse{Error: "not found"}}, nil
		}
		return oapi.DeleteEnv500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
	} else if env.TeamId != token.TeamId {
		return oapi.DeleteEnv404JSONResponse{NotFoundJSONResponse: oapi.NotFoundJSONResponse{Error: "not found"}}, nil
	}

	if err := a.deploymentStore.DeleteEnv(request.EnvId); err != nil {
		return oapi.DeleteEnv500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
	}

	return oapi.DeleteEnv204Response{}, nil
}

func (a api) GetEnv(ctx context.Context, request oapi.GetEnvRequestObject) (oapi.GetEnvResponseObject, error) {
	token := middleware.MustGetApiToken(ctx)

	env, err := a.deploymentStore.GetEnv(request.EnvId)
	if err != nil {
		if err == store.ErrEnvNotFound {
			return oapi.GetEnv404JSONResponse{NotFoundJSONResponse: oapi.NotFoundJSONResponse{Error: "not found"}}, nil
		}
		return oapi.GetEnv500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
	} else if env.TeamId != token.TeamId {
		return oapi.GetEnv404JSONResponse{NotFoundJSONResponse: oapi.NotFoundJSONResponse{Error: "not found"}}, nil
	}

	return oapi.GetEnv200JSONResponse(envFromStore(env)), nil
}

func (a api) CreateEnv(ctx context.Context, request oapi.CreateEnvRequestObject) (oapi.CreateEnvResponseObject, error) {
	token := middleware.MustGetApiToken(ctx)

	tid, err := typeid.FromString(request.EnvId)
	if err != nil {
		return oapi.CreateEnv400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: err.Error()}}, nil
	} else if tid.Prefix() != "env" {
		return oapi.CreateEnv400JSONResponse{BadRequestJSONResponse: oapi.BadRequestJSONResponse{Error: "invalid env id"}}, nil
	}

	env, err := a.deploymentStore.CreateEnv(store.CreateEnvOptions{
		TeamId: token.TeamId,
		Name:   request.Body.Name,
	})
	if err != nil {
		return oapi.CreateEnv500JSONResponse{InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{Error: err.Error()}}, nil
	}

	return oapi.CreateEnv201JSONResponse(envFromStore(env)), nil
}
