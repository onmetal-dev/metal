// package urls contains URL patterns for the app
package urls

import "fmt"

type Url interface {
	Pattern() string
	Render() string
}

type Static struct {
	Path string
}

var _ Url = Static{}

func (u Static) Pattern() string {
	return u.Path
}

func (u Static) Render() string {
	return u.Path
}

var Logout = Static{Path: "/logout"}
var Waitlist = Static{Path: "/waitlist"}
var Signup = Static{Path: "/signup"}
var Login = Static{Path: "/login"}
var About = Static{Path: "/about"}
var Health = Static{Path: "/health"}
var Onboarding = Static{Path: "/onboarding"}

type OnboardingPayment struct {
	TeamId string
}

var _ Url = OnboardingPayment{}

func (u OnboardingPayment) Pattern() string {
	return "/onboarding/{teamId}/payment"
}

func (u OnboardingPayment) Render() string {
	if u.TeamId == "" {
		panic("teamId is required")
	}
	return fmt.Sprintf("/onboarding/%s/payment", u.TeamId)
}

type OnboardingPaymentConfirm struct {
	TeamId string
}

var _ Url = OnboardingPaymentConfirm{}

func (u OnboardingPaymentConfirm) Pattern() string {
	return "/onboarding/{teamId}/payment/confirm"
}

func (u OnboardingPaymentConfirm) Render() string {
	if u.TeamId == "" {
		panic("teamId is required")
	}
	return fmt.Sprintf("/onboarding/%s/payment/confirm", u.TeamId)
}

type Home struct {
	TeamId  string
	EnvName string
}

const DefaultEnvSentinel = "x"

var _ Url = Home{}

func (u Home) Pattern() string {
	return "/dashboard/{teamId}/env/{envName}"
}

func (u Home) Render() string {
	if u.TeamId == "" || u.EnvName == "" {
		panic("teamId and envName are required")
	}
	return fmt.Sprintf("/dashboard/%s/env/%s", u.TeamId, u.EnvName)
}

type HomeSse struct {
	TeamId  string
	EnvName string
}

var _ Url = HomeSse{}

func (u HomeSse) Pattern() string {
	return "/dashboard/{teamId}/env/{envName}/sse"
}

func (u HomeSse) Render() string {
	if u.TeamId == "" || u.EnvName == "" {
		panic("teamId and envName are required")
	}
	return fmt.Sprintf("/dashboard/%s/env/%s/sse", u.TeamId, u.EnvName)
}

type NewServer struct {
	TeamId string
}

var _ Url = NewServer{}

func (u NewServer) Pattern() string {
	return "/dashboard/{teamId}/servers/new"
}

func (u NewServer) Render() string {
	if u.TeamId == "" {
		panic("teamId is required")
	}
	return fmt.Sprintf("/dashboard/%s/servers/new", u.TeamId)
}

type App struct {
	TeamId string
	AppId  string
}

var _ Url = App{}

func (u App) Pattern() string {
	return "/dashboard/{teamId}/apps/{appId}"
}

func (u App) Render() string {
	if u.TeamId == "" || u.AppId == "" {
		panic("teamId and appId are required")
	}
	return fmt.Sprintf("/dashboard/%s/apps/%s", u.TeamId, u.AppId)
}

type NewApp struct {
	TeamId string
}

var _ Url = NewApp{}

func (u NewApp) Pattern() string {
	return "/dashboard/{teamId}/apps/new"
}

func (u NewApp) Render() string {
	if u.TeamId == "" {
		panic("teamId is required")
	}
	return fmt.Sprintf("/dashboard/%s/apps/new", u.TeamId)
}

type DeploymentLogs struct {
	TeamId       string
	AppId        string
	EnvId        string
	DeploymentId uint
}

var _ Url = DeploymentLogs{}

func (u DeploymentLogs) Pattern() string {
	return "/dashboard/{teamId}/apps/{appId}/envs/{envId}/deployments/{deploymentId}/logs"
}

func (u DeploymentLogs) Render() string {
	if u.TeamId == "" || u.AppId == "" || u.EnvId == "" || u.DeploymentId == 0 {
		panic("teamId, appId, envId, and deploymentId are required")
	}
	return fmt.Sprintf("/dashboard/%s/apps/%s/envs/%s/deployments/%d/logs", u.TeamId, u.AppId, u.EnvId, u.DeploymentId)
}

type ServerCheckout struct {
	TeamId     string
	OfferingId string
	LocationId string
}

var _ Url = ServerCheckout{}

func (u ServerCheckout) Pattern() string {
	return "/dashboard/{teamId}/servers/checkout"
}

func (u ServerCheckout) Render() string {
	if u.TeamId == "" || u.OfferingId == "" || u.LocationId == "" {
		panic("teamId, offeringId, and locationId are required")
	}
	return fmt.Sprintf("/dashboard/%s/servers/checkout?offeringId=%s&locationId=%s", u.TeamId, u.OfferingId, u.LocationId)
}

type ServerCheckoutReturnUrl struct {
	TeamId string
}

var _ Url = ServerCheckoutReturnUrl{}

func (u ServerCheckoutReturnUrl) Pattern() string {
	return "/dashboard/{teamId}/servers/checkout-return-url"
}

func (u ServerCheckoutReturnUrl) Render() string {
	if u.TeamId == "" {
		panic("teamId is required")
	}
	return fmt.Sprintf("/dashboard/%s/servers/checkout-return-url", u.TeamId)
}

type TeamSettings struct {
	TeamId string
}

var _ Url = TeamSettings{}

func (u TeamSettings) Pattern() string {
	return "/dashboard/{teamId}/settings"
}

func (u TeamSettings) Render() string {
	if u.TeamId == "" {
		panic("teamId is required")
	}
	return fmt.Sprintf("/dashboard/%s/settings", u.TeamId)
}

type TeamInvites struct {
	TeamId string
}

var _ Url = TeamInvites{}

func (u TeamInvites) Pattern() string {
	return "/dashboard/{teamId}/invites"
}

func (u TeamInvites) Render() string {
	if u.TeamId == "" {
		panic("teamId is required")
	}
	return fmt.Sprintf("/dashboard/%s/invites", u.TeamId)
}

type DeleteTeamInvite struct {
	TeamId string
	Email  string
}

var _ Url = DeleteTeamInvite{}

func (u DeleteTeamInvite) Pattern() string {
	return "/dashboard/{teamId}/invites/{email}"
}

func (u DeleteTeamInvite) Render() string {
	if u.TeamId == "" || u.Email == "" {
		panic("teamId and email are required")
	}
	return fmt.Sprintf("/dashboard/%s/invites/%s", u.TeamId, u.Email)
}

type TeamApiTokens struct {
	TeamId string
}

var _ Url = TeamApiTokens{}

func (u TeamApiTokens) Pattern() string {
	return "/dashboard/{teamId}/apitokens"
}

func (u TeamApiTokens) Render() string {
	if u.TeamId == "" {
		panic("teamId is required")
	}
	return fmt.Sprintf("/dashboard/%s/apitokens", u.TeamId)
}

type DeleteTeamApiToken struct {
	TeamId     string
	ApiTokenId string
}

var _ Url = DeleteTeamApiToken{}

func (u DeleteTeamApiToken) Pattern() string {
	return "/dashboard/{teamId}/apitokens/{apiTokenId}"
}

func (u DeleteTeamApiToken) Render() string {
	if u.TeamId == "" || u.ApiTokenId == "" {
		panic("teamId and apiTokenId are required")
	}
	return fmt.Sprintf("/dashboard/%s/apitokens/%s", u.TeamId, u.ApiTokenId)
}

type EnvApp struct {
	TeamId  string
	AppId   string
	EnvName string
}

var _ Url = EnvApp{}

func (u EnvApp) Pattern() string {
	return "/dashboard/{teamId}/envs/{envName}/apps/{appId}"
}

func (u EnvApp) Render() string {
	if u.TeamId == "" || u.AppId == "" || u.EnvName == "" {
		panic("teamId, appId, and envName are required")
	}
	return fmt.Sprintf("/dashboard/%s/envs/%s/apps/%s", u.TeamId, u.EnvName, u.AppId)
}

type EnvAppDeployments struct {
	TeamId  string
	AppId   string
	EnvName string
}

var _ Url = EnvAppDeployments{}

func (u EnvAppDeployments) Pattern() string {
	return "/dashboard/{teamId}/envs/{envName}/apps/{appId}/deployments"
}

func (u EnvAppDeployments) Render() string {
	if u.TeamId == "" || u.AppId == "" || u.EnvName == "" {
		panic(fmt.Sprintf("teamId, appId, and envName are required: %v", u))
	}
	return fmt.Sprintf("/dashboard/%s/envs/%s/apps/%s/deployments", u.TeamId, u.EnvName, u.AppId)
}

type EnvAppVariables struct {
	TeamId  string
	AppId   string
	EnvName string
}

var _ Url = EnvAppVariables{}

func (u EnvAppVariables) Pattern() string {
	return "/dashboard/{teamId}/envs/{envName}/apps/{appId}/variables"
}

func (u EnvAppVariables) Render() string {
	if u.TeamId == "" || u.AppId == "" || u.EnvName == "" {
		panic("teamId, appId, and envName are required")
	}
	return fmt.Sprintf("/dashboard/%s/envs/%s/apps/%s/variables", u.TeamId, u.EnvName, u.AppId)
}

type EnvAppSettings struct {
	TeamId  string
	AppId   string
	EnvName string
}

var _ Url = EnvAppSettings{}

func (u EnvAppSettings) Pattern() string {
	return "/dashboard/{teamId}/envs/{envName}/apps/{appId}/settings"
}

func (u EnvAppSettings) Render() string {
	if u.TeamId == "" || u.AppId == "" || u.EnvName == "" {
		panic("teamId, appId, and envName are required")
	}
	return fmt.Sprintf("/dashboard/%s/envs/%s/apps/%s/settings", u.TeamId, u.EnvName, u.AppId)
}
