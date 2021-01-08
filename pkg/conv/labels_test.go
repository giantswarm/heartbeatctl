package conv_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"
	"github.com/opsgenie/opsgenie-go-sdk-v2/og"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/giantswarm/heartbeatctl/pkg/conv"
)

var _ = Describe("Labels", func() {
	Describe("HeartbeatAsLabels", func() {
		It("exposes Heartbeat fields as a Set of labels", func() {
			Expect(conv.HeartbeatAsLabels(heartbeat.Heartbeat{
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
			})).To(Equal(labels.Set{
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
			}))
		})

		It("treats false bool fields as absent labels", func() {
			Expect(conv.HeartbeatAsLabels(heartbeat.Heartbeat{
				Name:         "bar",
				Description:  "Heartbeat for bar",
				Interval:     4,
				IntervalUnit: "minutes",
				Enabled:      false,
				Expired:      false,
			})).To(Equal(labels.Set{
				"name":           "bar",
				"description":    "Heartbeat for bar",
				"interval":       "4",
				"intervalUnit":   "minutes",
				"ownerTeam/id":   "",
				"ownerTeam/name": "",
				"alertPriority":  "",
				"alertMessage":   "",
			}))
		})

		It("exposes unset fields as zero-values", func() {
			Expect(conv.HeartbeatAsLabels(heartbeat.Heartbeat{
				Name: "baz",
			})).To(Equal(labels.Set{
				"name":           "baz",
				"description":    "",
				"interval":       "0",
				"intervalUnit":   "",
				"ownerTeam/id":   "",
				"ownerTeam/name": "",
				"alertPriority":  "",
				"alertMessage":   "",
			}))
		})

		It("exposes tags as bool enabled labels but avoids overriding fields", func() {
			Expect(conv.HeartbeatAsLabels(heartbeat.Heartbeat{
				Name:         "qux",
				Description:  "Heartbeat for qux",
				Interval:     7,
				IntervalUnit: "minutes",
				Enabled:      true,
				AlertTags:    []string{"tagged", "not-overriding", "intervalUnit"},
			})).To(Equal(labels.Set{
				"name":           "qux",
				"description":    "Heartbeat for qux",
				"interval":       "7",
				"intervalUnit":   "minutes",
				"enabled":        "true",
				"ownerTeam/id":   "",
				"ownerTeam/name": "",
				"alertPriority":  "",
				"alertMessage":   "",
				"tagged":         "true",
				"not-overriding": "true",
			}))
		})

		It("parses tags structured as key:value", func() {
			Expect(conv.HeartbeatAsLabels(heartbeat.Heartbeat{
				Name:         "quux",
				IntervalUnit: "minutes",
				Enabled:      true,
				AlertTags:    []string{"tagged", "managed-by: foobricator", "intervalUnit:seconds", "panic-factor:7"},
			})).To(Equal(labels.Set{
				"name":           "quux",
				"description":    "",
				"interval":       "0",
				"intervalUnit":   "minutes",
				"enabled":        "true",
				"ownerTeam/id":   "",
				"ownerTeam/name": "",
				"alertPriority":  "",
				"alertMessage":   "",
				"tagged":         "true",
				"managed-by":     "foobricator",
				"panic-factor":   "7",
			}))
		})
	})
})
