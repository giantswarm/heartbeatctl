package cmdutil

import (
	"github.com/spf13/cobra"

	"github.com/giantswarm/heartbeatctl/pkg/ctl"
)

// SelectorOptions holds values for selectors/filtering options given on CLI
// and provides methods to configure a Cobra command instance with necessary
// flags, as well to transform options into a `ctl.SelectorConfig` suitable for
// using with ctl methods.
type SelectorOptions struct {
	nameExpressions []string
	labelSelector   string
	fieldSelector   string

	captureArgsUsingValidator bool
}

func NewSelectorOptions() *SelectorOptions {
	return &SelectorOptions{captureArgsUsingValidator: false}
}

// WithCapturingArgsUsingValidator configures this SelectorOptions
// (specifically its AddToCommand method) to also add hooks to given cobra
// Command to capture positional arguments and assign them as name expressions.
//
// It does this by assigning a function to `cmd.Args` that acts as a validator
// of received parameters, and also captures them and saves in SelectorOptions.
//
// Use this only when the command doesn't use its own positional arguments.
// If it does you can instead process the received parameters and separate
// the command's args from name expressions, and call NameExpressions method
// to assign them before a call to ToConfig.
func (so *SelectorOptions) WithCapturingArgsUsingValidator() *SelectorOptions {
	so.captureArgsUsingValidator = true
	return so
}

// AddFlags adds label and field selector flags to given cobra command.
// If capturing arguments was also previously enabled with a call to
// WithCapturingArgsUsingValidator, this will also add hook to the command that
// captures positional arguments and assigns them as name expressions in the
// selector.
func (so *SelectorOptions) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(
		&so.labelSelector, "selector", "l", so.labelSelector,
		"Selector (label query) to filter characters, supports '=', '==', '!=', 'in', 'notin', 'X', '!X'.",
	)
	flags.StringVar(
		&so.fieldSelector, "field-selector", so.fieldSelector,
		"Selector (field query) to filter characters, supports '=', '==', '!='.",
	)

	if !so.captureArgsUsingValidator {
		return
	}

	cmd.Args = so.argsCapturingValidator()
}

// NameExpressions assigns given names as name expressions to use for filtering
// objects by name.
func (so *SelectorOptions) NameExpressions(names ...string) *SelectorOptions {
	so.nameExpressions = names
	return so
}

// ToConfig takes values populated by CLI flags and produces a `SelectorConfig`
// that can be used with `ctl` app Port methods.
func (so *SelectorOptions) ToConfig() *ctl.SelectorConfig {
	return &ctl.SelectorConfig{
		NameExpressions: so.nameExpressions,
		LabelSelector:   so.labelSelector,
		FieldSelector:   so.fieldSelector,
	}
}

func (so *SelectorOptions) argsCapturingValidator() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			so.nameExpressions = args
		}
		return nil
	}
}
