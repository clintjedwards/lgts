package config

import (
	"github.com/kelseyhightower/envconfig"
)

//Config refers to general application configuration
type Config struct {
	ServerURL string `envconfig:"server_url" default:"localhost:8080"`
	Debug     bool   `envconfig:"debug" default:"false"`
	Database  *DatabaseConfig
	Slack     *SlackConfig
}

//DatabaseConfig refers to database connection settings
type DatabaseConfig struct {
	Name     string `envconfig:"database_name" default:"lgts"`
	URL      string `envconfig:"database_url" default:"localhost:5432"`
	User     string `envconfig:"database_user" default:"lgts"`
	Password string `envconfig:"database_password" default:"mysupersecretdbpassword"`
}

//SlackConfig refers to slack chat connection settings
type SlackConfig struct {
	AppToken string `envconfig:"slack_app_token" required:"true"`
	BotToken string `envconfig:"slack_bot_token" required:"true"`
}

//FromEnv pulls configuration from environment variables
func FromEnv() (*Config, error) {

	var config Config
	err := envconfig.Process("pollinate", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil

}
