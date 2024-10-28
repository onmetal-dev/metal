package up

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/onmetal-dev/metal/lib/cli/common"
	"github.com/onmetal-dev/metal/lib/cli/style"
	"github.com/onmetal-dev/metal/lib/cli/whoami"
	"github.com/onmetal-dev/metal/lib/oapi"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var docStyle = lipgloss.NewStyle().Margin(10, 2)
var textStyle = lipgloss.NewStyle().Foreground(style.BaseLight)

// item fulfills the list.Item interface required by list.Model
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

// getAppsMsg is a message sent when apps have been fetched from the API
type getAppsMsg struct {
	Apps  []oapi.App
	Error error
}

// getAppsCmd is a command that fetches apps from the API
// commands in bubbletea "make things happen" (i.e. use them for I/O) and emit messages that
// get sent to the model update function
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

// envsToItems converts a list of envs to a list of list.Items
func envsToItems(envs []oapi.Env) []list.Item {
	items := lo.Map(envs, func(env oapi.Env, _ int) list.Item {
		return item{
			title: env.Name,
			desc:  env.CreatedAt.Format(time.RFC3339),
		}
	})
	return items
}

// getEnvsMsg is a message sent when envs have been fetched from the API
type getEnvsMsg struct {
	Envs  []oapi.Env
	Error error
}

// getEnvsCmd is a command that fetches envs from the API
func getEnvsCmd(client oapi.ClientWithResponsesInterface) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetEnvsWithResponse(context.Background())
		if err != nil {
			return getEnvsMsg{Error: fmt.Errorf("error making request: %w", err)}
		}
		return getEnvsMsg{Envs: *resp.JSON200}
	}
}

// upRequestMsg is the result of initiating a request to the /up API endpoint
type upRequestMsg struct {
	Result io.ReadCloser
}

// upRequestCmd initiates a request to the /up API endpoint
func upRequestCmd(path string, part io.Writer, client oapi.ClientInterface, writer *multipart.Writer, body io.Reader) tea.Cmd {
	return func() tea.Msg {
		tgz, err := NewDirTargzipper(path, part)
		if err != nil {
			return upRequestIterMsg{Error: fmt.Errorf("error creating targzipper: %w", err)}
		}

		next, stop := iter.Pull2(tgz.Run())
		return upRequestIterMsg{Pull: iterPull2[Progress]{next: next, stop: stop}, Client: client, Writer: writer, Body: body}
	}
}

// iterPull2 captures the output of iter.Pull2 aka a pull-based iterator: https://tip.golang.org/blog/range-functions#pull-iterators
type iterPull2[T any] struct {
	next func() (T, error, bool)
	stop func()
}

// upRequestIterMsg captures the upload iterator's progress and keeps some other state (client, writer, body) around for when the iterator is done
type upRequestIterMsg struct {
	Error    error
	Progress *Progress
	Pull     iterPull2[Progress]
	Client   oapi.ClientInterface
	Writer   *multipart.Writer
	Body     io.Reader
}

// upRequestIterCmd wraps the iterator that tars up the app directory and produces progress updates in the form of upRequestIterMsg
func upRequestIterCmd(pull iterPull2[Progress], client oapi.ClientInterface, writer *multipart.Writer, body io.Reader) tea.Cmd {
	return func() tea.Msg {
		progress, err, ok := pull.next()
		if err != nil {
			return upRequestIterMsg{Error: fmt.Errorf("error from tgz iter: %w", err)}
		} else if ok {
			return upRequestIterMsg{
				Progress: &progress,
				Pull:     pull,
				Client:   client,
				Writer:   writer,
				Body:     body,
			}
		}
		pull.stop()
		if err := writer.Close(); err != nil {
			return upRequestIterMsg{Error: fmt.Errorf("error closing writer: %w", err)}
		}
		resp, err := client.UpWithBody(context.Background(), writer.FormDataContentType(), body)
		if err != nil {
			return upRequestIterMsg{Error: fmt.Errorf("error making request: %w", err)}
		} else if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return upRequestIterMsg{Error: fmt.Errorf("error reading response body: %w", err)}
			}
			return upRequestIterMsg{Error: fmt.Errorf("API returned non-200 status: %d: %s", resp.StatusCode, string(body))}
		}
		return upRequestMsg{Result: resp.Body}
	}
}

// upResponseMsg is sent when a line of the build / deploy logs is received
type upResponseMsg struct {
	Line    string
	Scanner *bufio.Scanner
	Error   error
	Done    bool
}

// streamUpResponse is a command that streams the build / deploy logs from the /up API endpoint
func streamUpResponse(scanner *bufio.Scanner) tea.Cmd {
	return func() tea.Msg {
		if scanner.Scan() {
			line := scanner.Text()
			return upResponseMsg{Scanner: scanner, Line: line}
		}
		if err := scanner.Err(); err != nil {
			return upResponseMsg{Done: true, Error: fmt.Errorf("error streaming up response: %w", err)}
		}
		return upResponseMsg{Done: true}
	}
}

var dump io.Writer

type model struct {
	flags     flags
	args      args
	exitError error

	width, height int
	loading       spinner.Model
	apiClient     oapi.ClientWithResponsesInterface
	apiClientRaw  oapi.ClientInterface
	authCheck     *whoami.Msg

	apps        *getAppsMsg
	selectedApp *oapi.App
	appList     *list.Model

	envs        *getEnvsMsg
	selectedEnv *oapi.Env
	envList     *list.Model

	upProgress   *progress.Model
	lastProgress *Progress
	upLogs       []string
	upDone       bool
	upError      error
}

var _ tea.Model = (*model)(nil)

func (m model) Init() tea.Cmd {
	return tea.Batch(m.loading.Tick, whoami.FetchWhoamiInfoCmd(m.apiClient))
}

const (
	createNewApp = "+ create a new app"
	createNewEnv = "+ create a new env"
)

const (
	padding  = 2
	maxWidth = 80
)

func finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*750, func(_ time.Time) tea.Msg {
		return nil
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if dump != nil {
		if m, ok := msg.(upRequestIterMsg); ok {
			spew.Fdump(dump, time.Now(), "upRequestIterMsg", map[string]any{"err": m.Error, "p": m.Progress})
		} else {
			spew.Fdump(dump, time.Now(), msg)
		}
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		h, v := docStyle.GetFrameSize()
		if m.appList != nil {
			m.appList.SetSize(msg.Width-h, msg.Height-v)
		} else if m.envList != nil {
			m.envList.SetSize(msg.Width-h, msg.Height-v)
		} else if m.upProgress != nil {
			m.upProgress.Width = msg.Width - padding*2 - 4
			if m.upProgress.Width > maxWidth {
				m.upProgress.Width = maxWidth
			}
		}
		return m, nil
	case spinner.TickMsg:
		if m.authCheck == nil {
			var cmd tea.Cmd
			m.loading, cmd = m.loading.Update(msg)
			return m, cmd
		}
	case whoami.Msg:
		m.authCheck = &msg
		if m.authCheck.Error == nil {
			return m, getAppsCmd(m.apiClient)
		}
		return m, nil
	case getAppsMsg:
		// list of apps received from API, render and wait for user to select one
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
		// list of envs received from API, render and wait for user to select one
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
	case upRequestIterMsg:
		if msg.Error != nil {
			m.exitError = fmt.Errorf("error uploading directory: %w", msg.Error)
			return m, tea.Quit
		}
		m.lastProgress = msg.Progress
		cmds := []tea.Cmd{upRequestIterCmd(msg.Pull, msg.Client, msg.Writer, msg.Body)}
		if msg.Progress != nil {
			cmds = append([]tea.Cmd{m.upProgress.SetPercent(msg.Progress.Percentage)}, cmds...)
		}
		return m, tea.Sequence(cmds...)
	case progress.FrameMsg:
		pm, cmd := m.upProgress.Update(msg)
		m.upProgress = lo.ToPtr(pm.(progress.Model))
		return m, cmd
	case upRequestMsg:
		//		m.up = &msg
		m.upLogs = []string{}
		return m, streamUpResponse(bufio.NewScanner(msg.Result))
	case upResponseMsg:
		if msg.Done {
			m.upDone = true
			m.upError = msg.Error
			return m, tea.Sequence(finalPause(), tea.Quit)
		}
		m.upLogs = append(m.upLogs, msg.Line)
		return m, streamUpResponse(msg.Scanner)
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
					m.upProgress = lo.ToPtr(progress.New(progress.WithGradient(string(style.Secondary), string(style.Primary))))
					var body bytes.Buffer
					writer := multipart.NewWriter(&body)
					if err := writer.WriteField("env_id", m.selectedEnv.Id); err != nil {
						m.exitError = fmt.Errorf("error writing env_id: %w", err)
						return m, tea.Quit
					}
					if err := writer.WriteField("app_id", m.selectedApp.Id); err != nil {
						m.exitError = fmt.Errorf("error writing app_id: %w", err)
						return m, tea.Quit
					}
					part, err := writer.CreateFormFile("archive", "archive.tar.gz")
					if err != nil {
						m.exitError = fmt.Errorf("error creating form file: %w", err)
						return m, tea.Quit
					}
					return m, upRequestCmd(m.args.path, part, m.apiClientRaw, writer, &body)
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

	// don't continue unless env is selected
	if m.selectedEnv == nil {
		return lipgloss.JoinVertical(lipgloss.Left, appSelection, envSelection)
	}

	// we have an app and env selected, time to upload the archive
	var upResult string
	if m.upProgress == nil {
		upResult = renderError(fmt.Errorf("unexpected nil upProgress"))
	} else if m.upProgress != nil {
		pad := strings.Repeat(" ", padding)
		verb := "uploading"
		file := ""
		if m.lastProgress != nil {
			if m.lastProgress.Done {
				verb = "uploaded"
				file = m.args.path + "!"
			} else {
				file = m.lastProgress.Filename + "..."
			}
		}
		upResult = lipgloss.JoinVertical(lipgloss.Left, textStyle.Render(fmt.Sprintf("%s %s", verb, file)), "\n"+
			pad+m.upProgress.View()+"\n\n")
	}

	var upLogs string
	if len(m.upLogs) > 0 {
		upLogs = strings.Join(m.upLogs, "\n")
		upLogs += "\n"
	}
	if m.upDone {
		if m.upError != nil {
			upLogs += textStyle.Render(fmt.Sprintf("deploy failed! ❌\n\n%s\n", m.upError))
		} else {
			upLogs += textStyle.Render("deploy completed! ✅\n")
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, appSelection, envSelection, upResult, upLogs)
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

// make this global so we can send messages to it from cmds
var p *tea.Program

func runUp(cmd *cobra.Command, argss []string) {
	if _, ok := os.LookupEnv("DEBUG"); ok {
		var err error
		dump, err = os.OpenFile("tmp/messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			os.Exit(1)
		}
	}

	path := "."
	if len(argss) > 0 {
		path = argss[0]
	}
	// convert path to absolute path
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		fmt.Println("error getting absolute path:", err)
		os.Exit(1)
	}

	p = tea.NewProgram(model{
		flags: flags{
			app: cmd.Flags().Lookup("app").Value.String(),
			env: cmd.Flags().Lookup("env").Value.String(),
		},
		args: args{
			path: path,
		},
		loading:      common.NewSpinner(),
		apiClient:    common.MustApiClient(),
		apiClientRaw: common.MustApiClientRaw(),
	})
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}
