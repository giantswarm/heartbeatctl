package cmd

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

const (
	StatusActive   = "ACTIVE"
	StatusDisabled = "DISABLED"
	StatusExpired  = "EXPIRED"
)

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List heartbeats",
		Run:   listRun,
	}

	status string
)

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&status, "status", "s", "", fmt.Sprintf("status of heartbeats to filter for, one of '%s', '%s', or '%s'", StatusActive, StatusDisabled, StatusExpired))
}

func getStatus(hb heartbeat.Heartbeat) string {
	if !hb.Enabled {
		return StatusDisabled
	}
	if hb.Expired {
		return StatusExpired
	}

	return StatusActive
}

func listRun(cmd *cobra.Command, args []string) {
	// Validate status flag.
	if status != "" && status != StatusActive && status != StatusDisabled && status != StatusExpired {
		log.Fatalf("status must be one of '%s', '%s', or '%s'\n", StatusActive, StatusDisabled, StatusExpired)
	}

	// List all heartbeats.
	result, err := heartbeatClient.List(context.Background())
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	listHeartbeats := result.Heartbeats

	// As expiry field is incorrect due to OpsGenie API bug,
	// request each heartbeat individually (in parallel),
	// which returns the correct expiry data.
	var wg sync.WaitGroup
	ch := make(chan heartbeat.Heartbeat)

	for _, hb := range listHeartbeats {
		wg.Add(1)

		go func(hb heartbeat.Heartbeat, ch chan heartbeat.Heartbeat) {
			newHb, err := heartbeatClient.Get(context.Background(), hb.Name)
			if err != nil {
				log.Fatalf("%v\n", err)
			}

			ch <- newHb.Heartbeat
		}(hb, ch)
	}

	heartbeats := []heartbeat.Heartbeat{}
	go func(wg *sync.WaitGroup) {
		for hb := range ch {
			heartbeats = append(heartbeats, hb)
			wg.Done()
		}
	}(&wg)

	wg.Wait()
	close(ch)

	// Filter by status.
	if status != "" {
		filteredHeartbeats := []heartbeat.Heartbeat{}
		for _, hb := range heartbeats {
			if status == getStatus(hb) {
				filteredHeartbeats = append(filteredHeartbeats, hb)
			}
		}
		heartbeats = filteredHeartbeats
	}

	// And sort, as we will have lost original ordering while requesting individual heartbeats.
	sort.Slice(heartbeats, func(i, j int) bool { return heartbeats[i].Name < heartbeats[j].Name })

	output := []string{}

	if !noHeaders {
		output = append(output, "NAME | STATUS")
	}

	for _, heartbeat := range heartbeats {
		output = append(output, fmt.Sprintf("%v | %v", heartbeat.Name, getStatus(heartbeat)))
	}

	fmt.Println(columnize.SimpleFormat(output))
}
