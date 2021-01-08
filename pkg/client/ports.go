package client

import (
	"context"

	"github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"
)

// This should be upstream in the API
// Method signatures taken from
// https://github.com/opsgenie/opsgenie-go-sdk-v2/blob/master/heartbeat/heartbeat.go

type Port interface {
	Ping(context context.Context, heartbeatName string) (*heartbeat.PingResult, error)
	Get(context context.Context, heartbeatName string) (*heartbeat.GetResult, error)
	List(context context.Context) (*heartbeat.ListResult, error)
	Update(context context.Context, request *heartbeat.UpdateRequest) (*heartbeat.HeartbeatInfo, error)
	Add(context context.Context, request *heartbeat.AddRequest) (*heartbeat.AddResult, error)
	Enable(context context.Context, heartbeatName string) (*heartbeat.HeartbeatInfo, error)
	Disable(context context.Context, heartbeatName string) (*heartbeat.HeartbeatInfo, error)
	Delete(context context.Context, heartbeatName string) (*heartbeat.DeleteResult, error)
}
