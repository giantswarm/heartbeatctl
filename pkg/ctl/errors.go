package ctl

import "errors"

// ErrNoSelector is an error returned when a particular method requires a
// non-empty selector but none was given.
var ErrNoSelector = errors.New(
	"no selector options given, to target all heartbeats pass '.*' name expression explicitly",
)
