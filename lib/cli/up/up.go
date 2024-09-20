package up

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/onmetal-dev/metal/lib/cli/common"
	"github.com/onmetal-dev/metal/lib/cli/style"
	"github.com/onmetal-dev/metal/lib/cli/whoami"
	"github.com/onmetal-dev/metal/lib/oapi"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var docStyle = lipgloss.NewStyle().Margin(10, 2)
var textStyle = lipgloss.NewStyle().Foreground(style.BaseLight)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func appsToItems(apps []oapi.App) []list.Item {
	items := lo.Map(apps, func(app oapi.App, _ int) list.Item {
		return item{
			title: app.Name,
			desc:  app.CreatedAt.Format(time.RFC3339),
		}
	})
	return items
}

type getAppsMsg struct {
	Apps  []oapi.App
	Error error
}

func getAppsCmd(client oapi.ClientWithResponsesInterface) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetAppsWithResponse(context.Background())
		if err != nil {
			return getAppsMsg{Error: fmt.Errorf("error making request: %w", err)}
		} else if resp.StatusCode() != http.StatusOK {
			return getAppsMsg{Error: fmt.Errorf("API returned non-200 status: %d: %s", resp.StatusCode(), string(resp.Body))}
		}
		return getAppsMsg{Apps: *resp.JSON200}
	}
}

func envsToItems(envs []oapi.Env) []list.Item {
	items := lo.Map(envs, func(env oapi.Env, _ int) list.Item {
		return item{
			title: env.Name,
			desc:  env.CreatedAt.Format(time.RFC3339),
		}
	})
	return items
}

type getEnvsMsg struct {
	Envs  []oapi.Env
	Error error
}

func getEnvsCmd(client oapi.ClientWithResponsesInterface) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetEnvsWithResponse(context.Background())
		if err != nil {
			return getEnvsMsg{Error: fmt.Errorf("error making request: %w", err)}
		}
		return getEnvsMsg{Envs: *resp.JSON200}
	}
}

type model struct {
	flags     flags
	args      args
	exitError error

	width, height int
	loading       spinner.Model
	apiClient     oapi.ClientWithResponsesInterface
	authCheck     *whoami.Msg

	apps        *getAppsMsg
	selectedApp *oapi.App
	appList     *list.Model

	envs        *getEnvsMsg
	selectedEnv *oapi.Env
	envList     *list.Model
}

var _ tea.Model = (*model)(nil)

func (m model) Init() tea.Cmd {
	return tea.Batch(m.loading.Tick, whoami.FetchWhoamiInfoCmd(m.apiClient))
}

const (
	createNewApp = "+ create a new app"
	createNewEnv = "+ create a new env"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		h, v := docStyle.GetFrameSize()
		if m.appList != nil {
			m.appList.SetSize(msg.Width-h, msg.Height-v)
		} else if m.envList != nil {
			m.envList.SetSize(msg.Width-h, msg.Height-v)
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.loading, cmd = m.loading.Update(msg)
		return m, cmd
	case whoami.Msg:
		m.authCheck = &msg
		if m.authCheck.Error == nil {
			return m, getAppsCmd(m.apiClient)
		}
		return m, nil
	case getAppsMsg:
		m.apps = &msg
		items := appsToItems(m.apps.Apps)
		items = append(items, item{
			title: createNewApp,
		})
		dd := list.NewDefaultDelegate()
		dd.ShowDescription = false
		dd.SetSpacing(0)
		appList := list.New(items, dd, 0, 0)
		appList.Title = "pick which app to deploy"
		appList.SetStatusBarItemName("option", "options")
		m.appList = &appList
		return m, nil
	case getEnvsMsg:
		m.envs = &msg
		items := envsToItems(m.envs.Envs)
		items = append(items, item{
			title: createNewEnv,
		})
		dd := list.NewDefaultDelegate()
		dd.ShowDescription = false
		dd.SetSpacing(0)
		envList := list.New(items, dd, 0, 0)
		envList.Title = "pick which env to deploy into"
		envList.SetStatusBarItemName("option", "options")
		m.envList = &envList
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.appList != nil && m.appList.SelectedItem() != nil {
				selected := m.appList.SelectedItem().(item)
				if selected.title == createNewApp {
					m.exitError = fmt.Errorf("TODO: not implemented")
					return m, tea.Quit
				}
				if app, ok := lo.Find(m.apps.Apps, func(app oapi.App) bool {
					return app.Name == selected.title
				}); !ok {
					m.exitError = fmt.Errorf("app %s not found", selected.title)
					return m, tea.Quit
				} else {
					m.selectedApp = &app
					m.appList = nil
					return m, getEnvsCmd(m.apiClient)
				}
			} else if m.envList != nil && m.envList.SelectedItem() != nil {
				selected := m.envList.SelectedItem().(item)
				if selected.title == createNewEnv {
					m.exitError = fmt.Errorf("TODO: not implemented")
					return m, tea.Quit
				}
				if env, ok := lo.Find(m.envs.Envs, func(env oapi.Env) bool {
					return env.Name == selected.title
				}); !ok {
					m.exitError = fmt.Errorf("env %s not found", selected.title)
					return m, tea.Quit
				} else {
					m.selectedEnv = &env
					m.envList = nil
					return m, tea.Quit
				}
			}
		}
	}

	var cmd tea.Cmd
	if m.appList != nil {
		*m.appList, cmd = m.appList.Update(msg)
	} else if m.envList != nil {
		*m.envList, cmd = m.envList.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	renderLoading := func(msg string, args ...any) string {
		return fmt.Sprintf("\n %s %s\n\n", m.loading.View(), textStyle.Render(fmt.Sprintf(msg, args...)))
	}
	renderError := func(err error) string {
		return fmt.Sprintf("%s\n", lipgloss.NewStyle().Foreground(style.Error).Render(fmt.Sprintf("error: %v", err)))
	}
	if m.exitError != nil {
		return renderError(m.exitError)
	}
	if m.authCheck == nil {
		return renderLoading("getting auth info...")
	}
	if m.authCheck.Error != nil {
		return renderError(m.authCheck.Error)
	}

	// we are auth'd. Next step is pulling down list of apps, potentially prompting user to select or create one
	var appSelection string
	if m.apps == nil {
		appSelection = renderLoading("getting list of apps...")
	} else if m.apps.Error != nil {
		appSelection = renderError(m.apps.Error)
	} else if m.selectedApp != nil {
		appSelection = textStyle.Render(fmt.Sprintf("app %s selected", m.selectedApp.Name))
	} else if m.appList == nil {
		appSelection = renderError(fmt.Errorf("unexpected nil appList after pulling apps down and with no selected app"))
	} else {
		m.appList.SetSize(m.width, m.height)
		appSelection = m.appList.View()
	}

	// don't continue unless app is selected
	if m.selectedApp == nil {
		return appSelection
	}

	// next step is pulling down list of envs, potentially prompting user to select or create one
	var envSelection string
	if m.envs == nil {
		envSelection = renderLoading("getting list of envs...")
	} else if m.envs.Error != nil {
		envSelection = renderError(m.envs.Error)
	} else if m.selectedEnv != nil {
		envSelection = textStyle.Render(fmt.Sprintf("env %s selected", m.selectedEnv.Name))
	} else if m.envList == nil {
		envSelection = renderError(fmt.Errorf("unexpected nil envList after pulling envs down"))
	} else {
		m.envList.SetSize(m.width, m.height-lipgloss.Height(appSelection))
		envSelection = m.envList.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, appSelection, envSelection)

}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "up [path]",
		Short:   "Launch an application. Defaults to launching the application code in the current directory.",
		Example: "metal up .",
		PreRun:  common.CheckToken,
		Run:     runUp,
		Args:    cobra.MaximumNArgs(1),
	}
	cmd.Flags().StringP("app", "a", "", "Specifies the app name to deploy. If not specified, will prompt interactively")
	cmd.Flags().StringP("env", "e", "", "Environment name to deploy into. If not specified, will prompt interactively")
	return cmd
}

type flags struct {
	app string
	env string
}

type args struct {
	path string
}

func runUp(cmd *cobra.Command, argss []string) {
	path := "."
	if len(argss) > 0 {
		path = argss[0]
	}
	p := tea.NewProgram(model{
		flags: flags{
			app: cmd.Flags().Lookup("app").Value.String(),
			env: cmd.Flags().Lookup("env").Value.String(),
		},
		args: args{
			path: path,
		},
		loading:   common.NewSpinner(),
		apiClient: common.MustApiClient(),
	})
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}
