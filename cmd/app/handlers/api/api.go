package api

import (
	"github.com/onmetal-dev/metal/lib/oapi"
	"github.com/onmetal-dev/metal/lib/store"
)

func New(
	apiTokenStore store.ApiTokenStore,
	appStore store.AppStore,
	deploymentStore store.DeploymentStore,
	teamStore store.TeamStore) oapi.StrictServerInterface {
	return api{
		apiTokenStore:   apiTokenStore,
		appStore:        appStore,
		deploymentStore: deploymentStore,
		teamStore:       teamStore,
	}
}

type api struct {
	apiTokenStore   store.ApiTokenStore
	appStore        store.AppStore
	deploymentStore store.DeploymentStore
	teamStore       store.TeamStore
}

var _ oapi.StrictServerInterface = api{}
