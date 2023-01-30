package conv_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"
	"github.com/opsgenie/opsgenie-go-sdk-v2/og"
	"k8s.io/apimachinery/pkg/fields"

	"github.com/giantswarm/heartbeatctl/pkg/conv"
)

var _ = Describe("Fields", func() {
	Describe("HeartbeatAsFields", func() {
		It("exposes Heartbeat fields as a Set of fields", func() {
			Expect(conv.HeartbeatAsFields(heartbeat.Heartbeat{
				Name:         "foo",
				Description:  "Heartbeat for foo",
				Interval:     5,
				IntervalUnit: "minutes",
				Enabled:      true,
				Expired:      true,
				OwnerTeam: og.OwnerTeam{
					Id:   "f000",
					Name: "a-team",
				},
				AlertPriority: "P2",
				AlertMessage:  "foo has no heartbeat",
			})).To(Equal(fields.Set{
				"name":           "foo",
				"description":    "Heartbeat for foo",
				"interval":       "5",
				"intervalUnit":   "minutes",
				"enabled":        "true",
				"expired":        "true",
				"ownerTeam/id":   "f000",
				"ownerTeam/name": "a-team",
				"alertPriority":  "P2",
				"alertMessage":   "foo has no heartbeat",
				"alertTags":      "",
			}))
		})

		It("treats false bool fields same as other fields", func() {
			Expect(conv.HeartbeatAsFields(heartbeat.Heartbeat{
				Name:         "bar",
				Description:  "Heartbeat for bar",
				Interval:     4,
				IntervalUnit: "minutes",
				Enabled:      false,
				Expired:      false,
			})).To(Equal(fields.Set{
				"name":           "bar",
				"description":    "Heartbeat for bar",
				"interval":       "4",
				"intervalUnit":   "minutes",
				"enabled":        "false",
				"expired":        "false",
				"ownerTeam/id":   "",
				"ownerTeam/name": "",
				"alertTags":      "",
				"alertPriority":  "",
				"alertMessage":   "",
			}))
		})

		It("exposes unset fields as zero-values", func() {
			Expect(conv.HeartbeatAsFields(heartbeat.Heartbeat{
				Name: "baz",
			})).To(Equal(fields.Set{
				"name":           "baz",
				"description":    "",
				"interval":       "0",
				"intervalUnit":   "",
				"enabled":        "false",
				"expired":        "false",
				"ownerTeam/id":   "",
				"ownerTeam/name": "",
				"alertTags":      "",
				"alertPriority":  "",
				"alertMessage":   "",
			}))
		})

		It("exposes tags as a field with comma-separated values", func() {
			Expect(conv.HeartbeatAsFields(heartbeat.Heartbeat{
				Name:         "qux",
				Description:  "Heartbeat for qux",
				Interval:     7,
				IntervalUnit: "minutes",
				Enabled:      false,
				AlertTags:    []string{"tagged", "not-overriding", "enabled", "intervalUnit"},
			})).To(Equal(fields.Set{
				"name":           "qux",
				"description":    "Heartbeat for qux",
				"interval":       "7",
				"intervalUnit":   "minutes",
				"enabled":        "false",
				"expired":        "false",
				"ownerTeam/id":   "",
				"ownerTeam/name": "",
				"alertTags":      "tagged,not-overriding,enabled,intervalUnit",
				"alertPriority":  "",
				"alertMessage":   "",
			}))
		})
	})
})
