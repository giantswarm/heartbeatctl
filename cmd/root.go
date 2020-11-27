package cmd

import (
	"log"
	"os"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	tokenEnvVar = "HEARTBEATCTL_TOKEN"
)

var (
	rootCmd = &cobra.Command{
		Use:   "heartbeatctl",
		Short: "heartbeatctl is a CLI tool to manage OpsGenie heartbeats",
	}

	noHeaders bool

	heartbeatClient *heartbeat.Client
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&noHeaders, "no-headers", false, "whether to disable headers")
}

func Execute() {
	token := os.Getenv(tokenEnvVar)
	if token == "" {
		log.Fatalf("%s cannot be empty\n", tokenEnvVar)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel)

	config := &client.Config{
		ApiKey: os.Getenv("HEARTBEATCTL_TOKEN"),
		Logger: logger,
	}

	c, err := heartbeat.NewClient(config)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	heartbeatClient = c

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%v\n", err)
	}
}
