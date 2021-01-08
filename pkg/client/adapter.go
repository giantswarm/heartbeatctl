package client

import (
	"fmt"
	"os"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"
	"github.com/sirupsen/logrus"
)

const (
	tokenEnvVar = "HEARTBEATCTL_TOKEN"
)

func New(cfg *client.Config) (Port, error) {
	if cfg == nil {
		cfg = &client.Config{}
	}

	if cfg.ApiKey == "" {
		token := os.Getenv(tokenEnvVar)
		if token == "" {
			return nil, fmt.Errorf("API key missing, set %s env var", tokenEnvVar)
		}
		cfg.ApiKey = token
	}

	if cfg.Logger == nil {
		cfg.Logger = logrus.New()
		cfg.Logger.SetLevel(logrus.PanicLevel)
	}

	return heartbeat.NewClient(cfg)
}
