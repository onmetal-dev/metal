package whoami

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/onmetal-dev/metal/lib/cli/style"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type apiConfig struct {
	baseUrl string
	token   string
}

type WhoamiResponse struct {
	TokenID   string `json:"token_id"`
	TeamID    string `json:"team_id"`
	TeamName  string `json:"team_name"`
	CreatedAt string `json:"created_at"`
}

type WhoamiMsg struct {
	Success *WhoamiResponse
	Error   error
}

type model struct {
	width, height int
	apiConfig     *apiConfig
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
	case apiConfig:
		m.apiConfig = &msg
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
		{"Token ID", whoami.TokenID},
		{"Team ID", whoami.TeamID},
		{"Team Name", whoami.TeamName},
		{"Token Created At", whoami.CreatedAt},
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
	go p.Send(apiConfig{
		baseUrl: viper.GetString("api-base-url"),
		token:   viper.GetString("api-token"),
	})
	_, err := p.Run()
	if err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}

func (m model) fetchWhoamiInfoCmd() tea.Msg {
	client := &http.Client{}
	req, err := http.NewRequest("GET", m.apiConfig.baseUrl+"/api/whoami", nil)
	if err != nil {
		return WhoamiMsg{Error: fmt.Errorf("error creating request: %w", err)}
	}

	req.Header.Set("Authorization", "Bearer "+m.apiConfig.token)

	resp, err := client.Do(req)
	if err != nil {
		return WhoamiMsg{Error: fmt.Errorf("error making request: %w", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WhoamiMsg{Error: fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)}
	}

	var whoamiResp WhoamiResponse
	if err := json.NewDecoder(resp.Body).Decode(&whoamiResp); err != nil {
		return WhoamiMsg{Error: fmt.Errorf("error decoding response: %w", err)}
	}

	return WhoamiMsg{Success: &whoamiResp}
}
