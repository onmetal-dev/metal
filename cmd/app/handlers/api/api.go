package api

import (
	"github.com/onmetal-dev/metal/lib/cellprovider"
	"github.com/onmetal-dev/metal/lib/oapi"
	"github.com/onmetal-dev/metal/lib/store"
)

func New(
	apiTokenStore store.ApiTokenStore,
	appStore store.AppStore,
	deploymentStore store.DeploymentStore,
	teamStore store.TeamStore,
	buildStore store.BuildStore,
	cellStore store.CellStore,
	cellProviderForType func(cellType store.CellType) cellprovider.CellProvider,
) oapi.StrictServerInterface {
	return api{
		apiTokenStore:       apiTokenStore,
		appStore:            appStore,
		deploymentStore:     deploymentStore,
		teamStore:           teamStore,
		buildStore:          buildStore,
		cellStore:           cellStore,
		cellProviderForType: cellProviderForType,
	}
}

type api struct {
	apiTokenStore       store.ApiTokenStore
	appStore            store.AppStore
	deploymentStore     store.DeploymentStore
	teamStore           store.TeamStore
	buildStore          store.BuildStore
	cellStore           store.CellStore
	cellProviderForType func(cellType store.CellType) cellprovider.CellProvider
}

var _ oapi.StrictServerInterface = api{}
