package conv

import (
	"fmt"
	"strings"

	"github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"
	"k8s.io/apimachinery/pkg/fields"
)

// HeartbeatAsFields transforms a Heartbeat struct into a field Set that
// implements the Fields interface.
func HeartbeatAsFields(h heartbeat.Heartbeat) fields.Set {
	return fields.Set{
		"name":           h.Name,
		"description":    h.Description,
		"interval":       fmt.Sprint(h.Interval),
		"enabled":        fmt.Sprint(h.Enabled),
		"intervalUnit":   h.IntervalUnit,
		"expired":        fmt.Sprint(h.Expired),
		"ownerTeam/id":   h.OwnerTeam.Id,
		"ownerTeam/name": h.OwnerTeam.Name,
		"alertTags":      strings.Join(h.AlertTags, ","),
		"alertPriority":  h.AlertPriority,
		"alertMessage":   h.AlertMessage,
	}
}
