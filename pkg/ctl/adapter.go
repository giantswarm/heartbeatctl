package ctl

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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
	ret, err := c.repo.List(context.Background())
	if err != nil {
		return nil, err
	}
	heartbeats := ret.Heartbeats

	if len(opts.NameExpressions) > 0 {
		heartbeats, err = filterNames(heartbeats, opts.NameExpressions)
		if err != nil {
			return nil, err
		}
	}

	// TODO: resolve heartbeats

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
	for _, h := range heartbeats {
		result, err := c.repo.Ping(context.Background(), h.Name)
		if err != nil {
			return pingResults, fmt.Errorf("heartbeat \"%s\" failed: %w", h.Name, err)
		}
		pingResults[h.Name] = *result
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
		hbi, err := meth(context.Background(), h.Name)
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
