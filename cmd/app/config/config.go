package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/onmetal-dev/metal/lib/validate"
)

type Config struct {
	Port                           string `envconfig:"PORT" default:"8080" required:"true"`
	SessionKey                     string `envconfig:"SESSION_KEY" default:"hownowbrowncow" required:"true"`
	SessionName                    string `envconfig:"SESSION_NAME" default:"session" required:"true"`
	DatabaseHost                   string `envconfig:"DATABASE_HOST" default:"localhost" required:"true"`
	DatabasePort                   int    `envconfig:"DATABASE_PORT" default:"5432" required:"true"`
	DatabaseUser                   string `envconfig:"DATABASE_USER" default:"postgres" required:"true"`
	DatabasePassword               string `envconfig:"DATABASE_PASSWORD" default:"postgres" required:"true"`
	DatabaseName                   string `envconfig:"DATABASE_NAME" default:"metal" required:"true"`
	DatabaseSslMode                string `envconfig:"DATABASE_SSL_MODE" default:"disable" required:"true"`
	LoopsWaitlistFormUrl           string `envconfig:"LOOPS_WAITLIST_FORM_URL" required:"true"`
	LoopsApiKey                    string `envconfig:"LOOPS_API_KEY" required:"true"`
	LoopsTxAddedToTeamNewUser      string `envconfig:"LOOPS_TX_ADDED_TO_TEAM_NEW_USER" required:"true"`
	LoopsTxAddedToTeamExistingUser string `envconfig:"LOOPS_TX_ADDED_TO_TEAM_EXISTING_USER" required:"true"`
	StripePublishableKey           string `envconfig:"STRIPE_PUBLISHABLE_KEY" required:"true"`
	StripeSecretKey                string `envconfig:"STRIPE_SECRET_KEY" required:"true"`
	HetznerRobotUsername           string `envconfig:"HETZNER_ROBOT_USERNAME" required:"true"`
	HetznerRobotPassword           string `envconfig:"HETZNER_ROBOT_PASSWORD" required:"true"`
	SshKeyBase64                   string `envconfig:"SSH_KEY_BASE64" required:"true"`
	SshKeyPassword                 string `envconfig:"SSH_KEY_PASSWORD" required:"true"`
	SshKeyFingerprint              string `envconfig:"SSH_KEY_FINGERPRINT" required:"true"`
	TmpDirRoot                     string `envconfig:"TMP_DIR_ROOT" default:"/tmp" required:"true"`
	CloudflareApiToken             string `envconfig:"CLOUDFLARE_API_TOKEN" required:"true"`
	CloudflareOnmetalDotRunZoneId  string `envconfig:"CLOUDFLARE_ONMETAL_DOT_RUN_ZONE_ID" required:"true"`
	Environment                    string `envconfig:"ENVIRONMENT" required:"true" validate:"oneof=local staging production"`
}

type Environment string

func (e Environment) IsProduction() bool {
	return e == EnvironmentProduction
}

const (
	EnvironmentLocal      Environment = "local"
	EnvironmentStaging    Environment = "staging"
	EnvironmentProduction Environment = "production"
)

// Set to local by default, can be overridden by loading config
var Env Environment = EnvironmentLocal

func loadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	if err := validate.Struct(&cfg); err != nil {
		return nil, err
	}
	Env = Environment(cfg.Environment)
	return &cfg, nil
}

func MustLoadConfig() *Config {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
