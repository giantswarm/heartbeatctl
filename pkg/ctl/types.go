package ctl

// SelectorConfig allow configuring selectors that specify field or label
// query expressions to filter a list of objects to operate on.
type SelectorConfig struct {
	// NameExpressions specifies a list of regular expressions to match against
	// names of heartbeats. Note that unlike label and field selectors
	// heartbeats matching any of the names are returned, but if both name
	// expressions and any selectors are specified at the same time, the
	// heartbeat must match all to be returned.
	NameExpressions []string

	// LabelSelector specifies a selector to filter the list of returned
	// objects using a K8s-compatible label selector syntax.
	// Defaults to accepting everything.
	LabelSelector string

	// FieldSelector specifies a selector to filter the list of returned
	// objects using a K8s-compatible field selector syntax.
	// Defaults to accepting everything.
	FieldSelector string
}

// empty returns true if all selector options are empty, i.e. selection space
// is not restricted and therefore all objects are targetted implicitly.
func (so *SelectorConfig) empty() bool {
	switch {
	case len(so.NameExpressions) > 0:
		return false
	case so.LabelSelector != "":
		return false
	case so.FieldSelector != "":
		return false
	default:
		return true
	}
}
