package conv

import (
	"fmt"
	"strings"

	"github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"
	"k8s.io/apimachinery/pkg/labels"
)

// HeartbeatAsLabels transforms a Heartbeat struct into a label Set that
// implements the Labels interface.
func HeartbeatAsLabels(h heartbeat.Heartbeat) labels.Labels {
	ls := labels.Set{
		"name":           h.Name,
		"description":    h.Description,
		"interval":       fmt.Sprint(h.Interval),
		"intervalUnit":   h.IntervalUnit,
		"ownerTeam/id":   h.OwnerTeam.Id,
		"ownerTeam/name": h.OwnerTeam.Name,
		"alertPriority":  h.AlertPriority,
		"alertMessage":   h.AlertMessage,
	}

	if h.Enabled {
		ls["enabled"] = fmt.Sprint(h.Enabled)
	}
	if h.Expired {
		ls["expired"] = fmt.Sprint(h.Expired)
	}

	for _, tag := range h.AlertTags {
		value := "true"
		if strings.Contains(tag, ":") {
			kv := strings.SplitN(tag, ":", 2)
			for i := range kv {
				kv[i] = strings.TrimSpace(kv[i])
			}
			tag, value = kv[0], kv[1]
		}

		if _, ok := ls[tag]; !ok {
			ls[tag] = value
		}
	}

	return ls
}
