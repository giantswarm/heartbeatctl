package cmd

import (
	"fmt"
	"log"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	"github.com/giantswarm/heartbeatctl/pkg/client"
	"github.com/giantswarm/heartbeatctl/pkg/cmdutil"
	"github.com/giantswarm/heartbeatctl/pkg/ctl"
)

// disableCmdOptions holds values for options accepted by the disable command
type disableCmdOptions struct {
	selectorOptions *cmdutil.SelectorOptions
}

var (
	disableDocLong = heredoc.Doc(`
		Disable specified heartbeats.

		Heartbeats to disable can be specified by a combination of selectors all of
		which must match against a heartbeat for it to be selected.

		The first is a kubectl-like label selector that can be specified using a
		'--selector' flag. Labels are any of the heartbeat fields like 'name',
		'interval', 'ownerTeam/name' or 'alertPriority', with boolean fields like
		'enabled' and 'expired' being present when the field is true and absent
		otherwise, and with additional labels generated from heartbeat's 'alertTags',
		so for example tag 'foo' becomes label 'foo' value 'true' and tag 'foo: bar'
		becomes label 'foo' value 'bar'. Label selector then allows you to use
		expressions with operators like '=', '==', '!=', 'in', 'notin', 'X', '!X'.
		
		The second is a kubectl-like field selector specified with '--field-selector'
		flag. This is similar but simpler, fields are exactly the same as fields in
		heartbeat object with first letter lowercased, and the only operators allowed
		are '=', '==', '!='.

		And finally any positional arguments are taken as regular expressions to match
		against heartbeat names. Multiple arguments can be given and they will be joined
		into an or-expression and wrapped in beginning and end-of-string bounds so the
		expressions have to match the entire name. E.g. parameters 'foo' 'bar-.*' will
		result in a regex '^(foo|bar-.*)$'.
	`)
	disableDocExamples = heredoc.Doc(`
		# disable all heartbeats with specified label 'managed-by' equal to 'foobricator'
		heartbeatctl disable --selector=managed-by=foobricator

		# disable all enabled heartbeats with label 'managed-by' equal to 'foobricator'
		heartbeatctl disable --selector="enabled,managed-by=foobricator"

		# disable all heartbeats with alert priority equal to 'P3'
		heartbeatctl disable --field-selector=alertPriority=P3

		# disable all non-expired heartbeats with alert priority equal to 'P2' or 'P4'
		heartbeatctl disable -l "!expired,alertPriority in (P2, P4)"

		# disable expired heartbeats with alert priority equal to 'P3'
		heartbeatctl disable -l expired --field-selector=alertPriority=P3

		# disable heartbeats with exact names
		heartbeatctl disable foo foo-rab1 bar-oof2

		# disable heartbeats with names matching any of the regular expressions
		heartbeatctl disable "foo.*" ".*-oof[12]"

		# disable enabled heartbeats with alert priority equal to 'P3' but only those
		# with names matching any of the given regular expressions
		heartbeatctl disable -l enabled --field-selector=alertPriority=P3 "foo.*" ".*-oof[12]"

		# disable all heartbeats (note that an explicit selector matching everything
		# must be given)
		heartbeatctl disable ".*"
	`)
)

func init() {
	rootCmd.AddCommand(NewCmdDisable())
}

func NewDisableOptions() *disableCmdOptions {
	return &disableCmdOptions{
		selectorOptions: cmdutil.NewSelectorOptions(),
	}
}

func NewCmdDisable() *cobra.Command {
	opts := NewDisableOptions()

	cmd := &cobra.Command{
		Use:     "disable [NAME..]",
		Short:   "Disable heartbeats",
		Long:    disableDocLong,
		Example: disableDocExamples,
		Run: func(cmd *cobra.Command, args []string) {
			runDisable(opts)
		},
	}

	opts.selectorOptions.WithCapturingArgsUsingValidator().AddFlags(cmd)

	return cmd
}

func runDisable(opts *disableCmdOptions) {
	repo, err := client.New(nil)
	if err != nil {
		log.Fatalf("Failed to init OpsGenie client: %v\n", err)
	}
	c := ctl.NewCtl(repo)

	heartbeats, err := c.Disable(opts.selectorOptions.ToConfig())
	for _, hbi := range heartbeats {
		fmt.Printf("heartbeat \"%s\" disabled\n", hbi.Name)
	}
	if err != nil {
		log.Fatalf("Failed to disable other heartbeats: %v\n", err)
	}
}
