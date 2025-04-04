package ctl_test

import (
	"errors"
	"fmt"
	"sort"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"

	"github.com/giantswarm/heartbeatctl/pkg/ctl"
	"github.com/giantswarm/heartbeatctl/pkg/mocks"
)

const (
	GetMethodName     = "Get"
	EnableMethodName  = "Enable"
	DisableMethodName = "Disable"
	PingMethodName    = "Ping"
)

func getName(h heartbeat.Heartbeat) string {
	return h.Name
}

// HeartbeatNamed returns a matcher that expects the Heartbeat to have `Name`
// attribute equal to specified value.
func HeartbeatNamed(name string) types.GomegaMatcher {
	return WithTransform(getName, Equal(name))
}

// ConsistOfHeartbeats returns a matcher that expects the value to `ConsistOf`
// only specified named heartbeats.
func ConsistOfHeartbeats(names ...string) types.GomegaMatcher {
	var matchers []interface{}
	for _, name := range names {
		matchers = append(matchers, HeartbeatNamed(name))
	}
	return ConsistOf(matchers...)
}

var _ = Describe("Adapter", func() {
	var (
		mockCtrl             *gomock.Controller
		repo                 *mocks.MockedClient
		adapter              ctl.Port
		configuredHeartbeats []heartbeat.Heartbeat
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		repo = mocks.NewMockedClient(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("success modes", func() {
		JustBeforeEach(func() {
			repo.EXPECT().List(gomock.Any()).Return(&heartbeat.ListResult{
				ResultMetadata: client.ResultMetadata{RequestId: "1", ResponseTime: 2.2},
				Heartbeats:     configuredHeartbeats,
			}, nil)
			for _, hb := range configuredHeartbeats {
				repo.EXPECT().Get(gomock.Any(), hb.Name).Return(&heartbeat.GetResult{
					ResultMetadata: client.ResultMetadata{RequestId: "1", ResponseTime: 2.2},
					Heartbeat:      hb,
				}, nil)
			}
			adapter = ctl.NewCtl(repo)
		})

		When("no heartbeats are configured", func() {
			BeforeEach(func() {
				configuredHeartbeats = []heartbeat.Heartbeat{}
			})

			Context(GetMethodName, func() {
				It("returns an empty list and no error", func() {
					Expect(adapter.Get(&ctl.SelectorConfig{})).To(BeEmpty())
				})

				It("ignores selector options", func() {
					Expect(adapter.Get(&ctl.SelectorConfig{
						LabelSelector: "name=foo",
					})).To(BeEmpty())
				})
			})
		})

		When("heartbeats are configured", func() {
			BeforeEach(func() {
				configuredHeartbeats = []heartbeat.Heartbeat{
					{
						Name:          "foo",
						Enabled:       true,
						Expired:       false,
						AlertPriority: "P2",
					},
					{
						Name:          "foo-oof1",
						Enabled:       true,
						Expired:       false,
						AlertPriority: "P3",
						AlertTags:     []string{"tagged", "managed-by: foobricator"},
					},
					{
						Name:          "foo-rab1",
						Enabled:       false,
						Expired:       false,
						AlertPriority: "P2",
						AlertTags:     []string{"tagged", "managed-by: foobricator"},
					},
					{
						Name:          "bar",
						Enabled:       true,
						Expired:       false,
						AlertPriority: "P2",
					},
					{
						Name:          "bar-oof2",
						Enabled:       true,
						Expired:       true,
						AlertPriority: "P3",
						AlertTags:     []string{"tagged", "managed-by: foobricator"},
					},
					{
						Name:          "bar-rab2",
						Enabled:       false,
						Expired:       false,
						AlertPriority: "P4",
						AlertTags:     []string{"tagged", "managed-by: foobricator"},
					},
				}
				sort.Slice(configuredHeartbeats, func(i, j int) bool {
					return configuredHeartbeats[i].Name < configuredHeartbeats[j].Name
				})
			})

			Context(GetMethodName, func() {
				DescribeTable(
					"filters elements correctly",
					func(opts *ctl.SelectorConfig, expected ...string) {
						Expect(adapter.Get(opts)).To(ConsistOfHeartbeats(expected...))
					},
					Entry(
						"returns everything without filters",
						&ctl.SelectorConfig{},
						"bar", "bar-oof2", "bar-rab2", "foo", "foo-oof1", "foo-rab1",
					),
					Entry(
						"returns only heartbeats matching a label selector when given",
						&ctl.SelectorConfig{LabelSelector: "managed-by=foobricator"},
						"bar-oof2", "bar-rab2", "foo-oof1", "foo-rab1",
					),
					Entry(
						"returns only heartbeats matching a complex label selector when given",
						&ctl.SelectorConfig{LabelSelector: "!enabled,managed-by=foobricator"},
						"bar-rab2", "foo-rab1",
					),
					Entry(
						"returns only heartbeats matching a field selector when given",
						&ctl.SelectorConfig{FieldSelector: "alertPriority=P3"},
						"bar-oof2", "foo-oof1",
					),
					Entry(
						"returns only heartbeats matching both label and field selectors when given",
						&ctl.SelectorConfig{LabelSelector: "expired", FieldSelector: "alertPriority=P3"},
						"bar-oof2",
					),
					Entry(
						"returns specific heartbeats when given explicit names",
						&ctl.SelectorConfig{NameExpressions: []string{"foo", "foo-rab1", "bar-oof2"}},
						"bar-oof2", "foo", "foo-rab1",
					),
					Entry(
						"returns a union of sets of heartbeats for any passed name regexes",
						&ctl.SelectorConfig{NameExpressions: []string{"foo.*", ".*-oof[12]"}},
						"bar-oof2", "foo", "foo-oof1", "foo-rab1",
					),
					Entry(
						"returns only heartbeats matching all selectors when given",
						&ctl.SelectorConfig{
							NameExpressions: []string{"foo.*", ".*-oof[12]"},
							LabelSelector:   "enabled",
							FieldSelector:   "alertPriority=P3",
						},
						"bar-oof2", "foo-oof1",
					),
				)
			})

			// AssertMethodCalledOnSelectedHeartbeats asserts method `$name` is
			// called on all heartbeats that should be matched by some selector.
			AssertMethodCalledOnSelectedHeartbeats := func(methodName string) {
				Context(methodName, func() {
					var (
						expected      []string
						expectedInfos []heartbeat.HeartbeatInfo
					)

					JustBeforeEach(func() {
						expected = []string{"bar-oof2", "foo-oof1"}
						for _, hbName := range expected {
							hbi := &heartbeat.HeartbeatInfo{
								Name:    hbName,
								Enabled: true,
								Expired: false,
							}

							// setup the right expectation, depending on
							// which method we're asserting
							switch methodName {
							case EnableMethodName:
								repo.EXPECT().Enable(gomock.Any(), hbName).Return(hbi, nil)
							case DisableMethodName:
								// in case of disable a successful call
								// would set `Enabled` to false
								hbi.Enabled = false
								repo.EXPECT().Disable(gomock.Any(), hbName).Return(hbi, nil)
							}

							expectedInfos = append(expectedInfos, *hbi)
						}
					})

					It(fmt.Sprintf("calls %s on heartbeats selected by given options", methodName), func() {
						method := adapter.Enable
						if methodName == DisableMethodName {
							method = adapter.Disable
						}

						Expect(method(&ctl.SelectorConfig{
							NameExpressions: []string{"foo.*", ".*-oof[12]"},
							LabelSelector:   "enabled",
							FieldSelector:   "alertPriority=P3",
						})).To(Equal(expectedInfos))
					})
				})
			}

			AssertMethodCalledOnSelectedHeartbeats(EnableMethodName)
			AssertMethodCalledOnSelectedHeartbeats(DisableMethodName)

			AssertMethodFailsFastWhenRepoCallFails := func(methodName string) {
				Context(methodName, func() {
					It("fails fast when first repo call on a heartbeat fails", func() {
						By("making first heartbeat succeed and second fail")

						fooHbi := &heartbeat.HeartbeatInfo{
							Name:    "foo",
							Enabled: true,
							Expired: false,
						}
						apiErr := errors.New("API call failed")

						switch methodName {
						case EnableMethodName:
							repo.EXPECT().Enable(gomock.Any(), "foo").Return(fooHbi, nil)
							repo.EXPECT().Enable(gomock.Any(), "foo-oof1").Return(nil, apiErr)
						case DisableMethodName:
							repo.EXPECT().Disable(gomock.Any(), "foo").Return(fooHbi, nil)
							repo.EXPECT().Disable(gomock.Any(), "foo-oof1").Return(nil, apiErr)
						}

						By("calling adapter method")

						method := adapter.Enable
						if methodName == DisableMethodName {
							method = adapter.Disable
						}
						hbInfos, err := method(&ctl.SelectorConfig{
							NameExpressions: []string{"foo.*"},
						})

						By("ensuring we get an error")

						Expect(err).To(SatisfyAll(
							MatchError(apiErr),
							// assert that error tells us which heartbeat caused the error
							WithTransform(
								func(e error) string { return e.Error() },
								ContainSubstring("foo-oof1"),
							),
						))

						By("ensuring we also get info about the heartbeat that succeeded")

						Expect(hbInfos).To(Equal([]heartbeat.HeartbeatInfo{*fooHbi}))
					})
				})
			}

			AssertMethodFailsFastWhenRepoCallFails(EnableMethodName)
			AssertMethodFailsFastWhenRepoCallFails(DisableMethodName)

			Context(PingMethodName, func() {
				var (
					expected      []string
					expectedInfos = make(map[string]heartbeat.PingResult)
				)

				JustBeforeEach(func() {
					expected = []string{"foo-oof1", "bar-oof2"}
					for _, hbName := range expected {
						ping := &heartbeat.PingResult{
							Message: "PONG - Heartbeat received",
						}

						repo.EXPECT().Ping(gomock.Any(), hbName).Return(ping, nil)
						expectedInfos[hbName] = *ping
					}
				})

				It("calls Ping on heartbeats selected by given options", func() {
					Expect(adapter.Ping(&ctl.SelectorConfig{
						NameExpressions: []string{"foo.*", ".*-oof[12]"},
						LabelSelector:   "enabled",
						FieldSelector:   "alertPriority=P3",
					})).To(Equal(expectedInfos))
				})
			})

			Context(PingMethodName, func() {
				It("fails fast when first repo call on a heartbeat fails", func() {
					By("making first heartbeat succeed and second fail")

					fooPingResult := &heartbeat.PingResult{
						Message: "PONG - Heartbeat received",
					}
					apiErr := errors.New("API call failed")

					repo.EXPECT().Ping(gomock.Any(), "foo-rab1").Return(fooPingResult, nil)
					repo.EXPECT().Ping(gomock.Any(), "foo").Return(fooPingResult, nil)
					repo.EXPECT().Ping(gomock.Any(), "foo-oof1").Return(nil, apiErr)

					By("calling adapter method")

					pingResults, err := adapter.Ping(&ctl.SelectorConfig{
						NameExpressions: []string{"foo.*"},
					})

					By("ensuring we get an error")

					Expect(err).To(SatisfyAll(
						MatchError(fmt.Errorf("heartbeats \"[foo-oof1]\" failed")),
						// assert that error tells us which heartbeat caused the error
						WithTransform(
							func(e error) string { return e.Error() },
							ContainSubstring("foo-oof1"),
						),
					))

					By("ensuring we also get info about the heartbeat that succeeded")

					Expect(pingResults).To(Equal(map[string]heartbeat.PingResult{
						"foo":      *fooPingResult,
						"foo-rab1": *fooPingResult,
					}))
				})
			})
		})

	})

	Describe("failure modes", func() {
		JustBeforeEach(func() {
			adapter = ctl.NewCtl(repo)
		})

		When("repo List returns an error", func() {
			var apiErr error

			JustBeforeEach(func() {
				apiErr = errors.New("API request failed")
				repo.EXPECT().List(gomock.Any()).Return(nil, apiErr)
			})

			AssertMethodPropagatesError := func(methodName string) {
				Context(methodName, func() {
					It("propagates the error", func() {
						var err error
						opts := &ctl.SelectorConfig{NameExpressions: []string{".*"}}

						switch methodName {
						case GetMethodName:
							_, err = adapter.Get(opts)
						case EnableMethodName:
							_, err = adapter.Enable(opts)
						case DisableMethodName:
							_, err = adapter.Disable(opts)
						case PingMethodName:
							_, err = adapter.Ping(opts)
						}

						Expect(err).NotTo(Succeed())
						Expect(err).To(MatchError(apiErr))
					})
				})
			}

			AssertMethodPropagatesError(GetMethodName)
			AssertMethodPropagatesError(EnableMethodName)
			AssertMethodPropagatesError(DisableMethodName)
			AssertMethodPropagatesError(PingMethodName)
		})

		When("no selectors are given", func() {
			var methods map[string]func(*ctl.SelectorConfig) ([]heartbeat.HeartbeatInfo, error)

			BeforeEach(func() {
				methods = map[string]func(*ctl.SelectorConfig) ([]heartbeat.HeartbeatInfo, error){
					EnableMethodName:  adapter.Enable,
					DisableMethodName: adapter.Disable,
				}
			})

			// AssertMethodFails is a shared behaviour that asserts one of the
			// pre-defined methods fails when no selector options were given.
			AssertMethodFails := func(name string) {
				Context(name, func() {
					It("fails", func() {
						hbInfos, err := methods[name](&ctl.SelectorConfig{})
						Expect(err).To(MatchError(
							"no selector options given, to target all heartbeats pass '.*' name expression explicitly",
						))
						Expect(hbInfos).To(BeNil())
					})
				})
			}

			AssertMethodFails(EnableMethodName)
			AssertMethodFails(DisableMethodName)

			Context(PingMethodName, func() {
				It("fails", func() {
					results, err := adapter.Ping(&ctl.SelectorConfig{})
					Expect(err).To(MatchError(
						"no selector options given, to target all heartbeats pass '.*' name expression explicitly",
					))
					Expect(results).To(BeNil())
				})
			})
		})
	})
})
