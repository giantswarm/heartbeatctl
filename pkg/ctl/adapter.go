package ctl

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/giantswarm/heartbeatctl/pkg/client"
	"github.com/giantswarm/heartbeatctl/pkg/conv"
)

type ctl struct {
	repo client.Port
}

func NewCtl(r client.Port) Port {
	return &ctl{repo: r}
}

func (c *ctl) Get(opts *SelectorConfig) ([]heartbeat.Heartbeat, error) {
	ret, err := c.repo.List(context.TODO())
	if err != nil {
		return nil, err
	}
	// As expiry field is incorrect due to OpsGenie API bug,
	// request each heartbeat individually (in parallel),
	// which returns the correct expiry data.
	var wg sync.WaitGroup
	ch := make(chan heartbeat.Heartbeat)

	for _, hb := range ret.Heartbeats {
		wg.Add(1)

		go func(hb heartbeat.Heartbeat, ch chan heartbeat.Heartbeat) {
			newHb, err := c.repo.Get(context.Background(), hb.Name)
			if err != nil {
				log.Fatalf("%v\n", err)
			}

			ch <- newHb.Heartbeat
		}(hb, ch)
	}

	heartbeats := make([]heartbeat.Heartbeat, 0)
	go func(wg *sync.WaitGroup) {
		for hb := range ch {
			heartbeats = append(heartbeats, hb)
			wg.Done()
		}
	}(&wg)

	wg.Wait()
	close(ch)

	if len(opts.NameExpressions) > 0 {
		heartbeats, err = filterNames(heartbeats, opts.NameExpressions)
		if err != nil {
			return nil, err
		}
	}

	ls := labels.Everything()
	if opts.LabelSelector != "" {
		ls, err = labels.Parse(opts.LabelSelector)
		if err != nil {
			return nil, err
		}
	}

	fs := fields.Everything()
	if opts.FieldSelector != "" {
		fs, err = fields.ParseSelector(opts.FieldSelector)
		if err != nil {
			return nil, err
		}
	}

	filtered := []heartbeat.Heartbeat{}
	for _, h := range heartbeats {
		if !ls.Matches(conv.HeartbeatAsLabels(h)) {
			continue
		}

		if !fs.Matches(conv.HeartbeatAsFields(h)) {
			continue
		}

		filtered = append(filtered, h)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})

	return filtered, nil
}

func (c *ctl) Enable(opts *SelectorConfig) ([]heartbeat.HeartbeatInfo, error) {
	return c.enableDisableHeartbeats(c.repo.Enable, opts)
}

func (c *ctl) Disable(opts *SelectorConfig) ([]heartbeat.HeartbeatInfo, error) {
	return c.enableDisableHeartbeats(c.repo.Disable, opts)
}

func (c *ctl) Ping(opts *SelectorConfig) (map[string]heartbeat.PingResult, error) {
	if opts.empty() {
		return nil, ErrNoSelector
	}

	heartbeats, err := c.Get(opts)
	if err != nil {
		return nil, err
	}

	var pingResults = make(map[string]heartbeat.PingResult)
	var failedPings = make([]string, 0)
	for _, h := range heartbeats {
		result, err := c.repo.Ping(context.Background(), h.Name)
		if err != nil {
			failedPings = append(failedPings, h.Name)
		}

		if result != nil {
			pingResults[h.Name] = *result
		}
	}

	if len(failedPings) > 0 {
		return pingResults, fmt.Errorf("heartbeats \"%v\" failed", failedPings)
	}
	return pingResults, nil
}

// enableDisableHeartbeats applies given method (can be either `repo.Enable` or
// `repo.Disable`) to all heartbeats matched by given selector options, which
// must be non-empty.
func (c *ctl) enableDisableHeartbeats(meth func(context.Context, string) (*heartbeat.HeartbeatInfo, error), opts *SelectorConfig) ([]heartbeat.HeartbeatInfo, error) {
	if opts.empty() {
		return nil, ErrNoSelector
	}

	heartbeats, err := c.Get(opts)
	if err != nil {
		return nil, err
	}

	var hbInfos []heartbeat.HeartbeatInfo
	for _, h := range heartbeats {
		// TODO: context.TODO
		hbi, err := meth(context.TODO(), h.Name)
		if err != nil {
			return hbInfos, fmt.Errorf("heartbeat \"%s\" failed: %w", h.Name, err)
		}
		hbInfos = append(hbInfos, *hbi)
	}
	return hbInfos, nil
}

func filterNames(heartbeats []heartbeat.Heartbeat, nameExpressions []string) ([]heartbeat.Heartbeat, error) {
	expr, err := regexp.Compile(fmt.Sprintf("^(%s)$", strings.Join(nameExpressions, "|")))
	if err != nil {
		return nil, err
	}

	filtered := []heartbeat.Heartbeat{}
	for _, h := range heartbeats {
		if expr.MatchString(h.Name) {
			filtered = append(filtered, h)
		}
	}
	return filtered, nil
}
