package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

var (
	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get heartbeat",
		Run:   getRun,
		Args:  cobra.ExactArgs(1),
	}
)

func init() {
	rootCmd.AddCommand(getCmd)
}

func getRun(cmd *cobra.Command, args []string) {
	heartbeat, err := heartbeatClient.Get(context.Background(), args[0])
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	output := []string{}

	if !noHeaders {
		output = append(output, "NAME | STATUS")
	}

	output = append(output, fmt.Sprintf("%v | %v", heartbeat.Name, getStatus(heartbeat.Heartbeat)))

	fmt.Println(columnize.SimpleFormat(output))
}
