package config

import (
	"github.com/kelseyhightower/envconfig"
)

//Config refers to general application configuration
type Config struct {
	ServerURL string `envconfig:"server_url" default:"localhost:8080"`
	Debug     bool   `envconfig:"debug" default:"false"`
	Slack     *SlackConfig
}

//SlackConfig refers to slack chat connection settings
type SlackConfig struct {
	AppToken string `envconfig:"slack_app_token" required:"true"`
	BotToken string `envconfig:"slack_bot_token" required:"true"`
}

//FromEnv pulls configuration from environment variables
func FromEnv() (*Config, error) {

	var config Config
	err := envconfig.Process("Snark", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil

}
