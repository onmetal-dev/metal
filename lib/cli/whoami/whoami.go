package whoami

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/onmetal-dev/metal/lib/cli/style"
	"github.com/onmetal-dev/metal/lib/oapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type WhoamiMsg struct {
	Success *oapi.WhoAmI
	Error   error
}

type model struct {
	width, height int
	apiClient     oapi.ClientWithResponsesInterface
	whoamiMsg     *WhoamiMsg
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case oapi.ClientWithResponsesInterface:
		m.apiClient = msg
		return m, m.fetchWhoamiInfoCmd
	case WhoamiMsg:
		m.whoamiMsg = &msg
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.whoamiMsg == nil {
		return "Loading..."
	}

	if m.whoamiMsg.Error != nil {
		return fmt.Sprintf("%s\n", lipgloss.NewStyle().Foreground(style.Error).Render(fmt.Sprintf("Error: %v", m.whoamiMsg.Error)))
	}

	whoami := m.whoamiMsg.Success
	rows := [][]string{
		{"Token ID", whoami.TokenId},
		{"Team ID", whoami.TeamId},
		{"Team Name", whoami.TeamName},
		{"Token Created At", whoami.CreatedAt.Format(time.RFC3339)},
	}

	baseStyle := lipgloss.NewStyle().Foreground(style.Primary)
	t := table.New().
		Width(m.width).
		Height(m.height).
		Border(lipgloss.NormalBorder()).
		BorderStyle(baseStyle).
		Headers("Field", "Value").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return baseStyle.Foreground(style.Neutral).Bold(true)
			}
			return baseStyle.Foreground(style.Neutral)
		}).
		Rows(rows...).
		Width(70)
	return t.Render() + "\n"
}

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Display information about the current API token",
		Run:   runWhoami,
	}
}

func runWhoami(cmd *cobra.Command, args []string) {
	p := tea.NewProgram(model{})
	client, err := oapi.NewClientWithResponses(viper.GetString("api-base-url"),
		oapi.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+viper.GetString("api-token"))
			return nil
		}))
	if err != nil {
		fmt.Println("error creating client:", err)
		os.Exit(1)
	}
	go p.Send(client)
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}

func (m model) fetchWhoamiInfoCmd() tea.Msg {
	resp, err := m.apiClient.WhoAmIWithResponse(context.Background())
	if err != nil {
		return WhoamiMsg{Error: fmt.Errorf("error making request: %w", err)}
	} else if resp.StatusCode() != http.StatusOK {
		return WhoamiMsg{Error: fmt.Errorf("API returned non-200 status: %d: %s", resp.StatusCode(), string(resp.Body))}
	}
	return WhoamiMsg{Success: resp.JSON200}
}
