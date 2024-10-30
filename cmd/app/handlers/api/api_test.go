package api

import "github.com/onmetal-dev/metal/lib/store/mock"

func newTestAPI() api {
	return New(
		&mock.ApiTokenStoreMock{},
		&mock.AppStoreMock{},
		&mock.DeploymentStoreMock{},
		&mock.TeamStoreMock{},
		nil,
		nil,
		nil,
	).(api)
}
