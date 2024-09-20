package common

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/onmetal-dev/metal/lib/cli/style"
	"github.com/onmetal-dev/metal/lib/oapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func CheckToken(cmd *cobra.Command, args []string) {
	token := viper.GetString("api-token")
	if token == "" {
		fmt.Println("Error: Token is not set. Please set it using --token flag or in the config file.")
		os.Exit(1)
	}
}

func NewSpinner() spinner.Model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(style.Primary)
	s.Spinner = spinner.Line
	return s
}

func MustApiClient() oapi.ClientWithResponsesInterface {
	client, err := oapi.NewClientWithResponses(viper.GetString("api-base-url"),
		oapi.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+viper.GetString("api-token"))
			return nil
		}))
	if err != nil {
		fmt.Println("error creating client:", err)
		os.Exit(1)
	}
	return client
}
