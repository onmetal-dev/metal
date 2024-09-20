package cli

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/onmetal-dev/metal/lib/cli/up"
	"github.com/onmetal-dev/metal/lib/cli/whoami"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "metal",
	Short: "Bare metal PaaS",
	Long: `Metal is platform for depoying applications on bare metal.
The CLI can be used to interact with the platform.

Certain flags to the CLI can be configured via environment variables or config file values. 
E.g., instead of --api-token you can use METAL_API_TOKEN or put api-token: ... in a config file.

The default location of the config file is $HOME/.metal/config.yaml, but can be overridden with the --config flag.

Precedence order: CLI flags > environment variables > config file > default value.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		apiToken := viper.GetString("api-token")
		if apiToken == "" {
			return fmt.Errorf("apitoken is not set. Please provide it via CLI flag or config file")
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.metal/config.yaml)")
	rootCmd.PersistentFlags().String("api-base-url", "https://www.onmetal.dev", "API base URL")
	rootCmd.PersistentFlags().String("api-token", "", "Token for authentication")
	rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		viper.BindPFlag(flag.Name, flag)
	})

	rootCmd.AddCommand(whoami.NewCmd())
	rootCmd.AddCommand(up.NewCmd())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".metal" (without extension).
		viper.AddConfigPath(path.Join(home, ".metal"))
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("METAL")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading config file:", err)
		os.Exit(1)
	}
}
