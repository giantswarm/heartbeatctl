package ctl

import "github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"

// Port of the heartbeatctl application
type Port interface {
	// Get returns a list of Heartbeats. If SelectorConfig is given, the list
	// is filtered down to only include Heartbeats matching both the label and
	// field selectors or name expressions (all of which are optional),
	// otherwise all Heartbeats are returned.
	Get(*SelectorConfig) ([]heartbeat.Heartbeat, error)

	// Enable enables all heartbeats selected by given SelectorConfig, which
	// in this case must specify at least one selector or name (to target all
	// heartbeats specify a `nameExpressions=['.*']` rule explicitly).
	Enable(*SelectorConfig) ([]heartbeat.HeartbeatInfo, error)

	// Disable disables all heartbeats selected by given SelectorConfig, which
	// in this case must specify at least one selector or name (to target all
	// heartbeats specify a `nameExpressions=['.*']` rule explicitly).
	Disable(*SelectorConfig) ([]heartbeat.HeartbeatInfo, error)

	// Ping pings all heartbeats selected by given SelectorConfig, which
	// in this case must specify at least one selector or name (to target all
	// heartbeats specify a `nameExpressions=['.*']` rule explicitly).
	Ping(*SelectorConfig) (map[string]heartbeat.PingResult, error)
}
