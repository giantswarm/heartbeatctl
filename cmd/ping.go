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

// pingCmdOptions holds values for options accepted by the ping command
type pingCmdOptions struct {
	selectorOptions *cmdutil.SelectorOptions
}

var (
	pingDocLong = heredoc.Doc(`
	        Issue a ping request to specified heartbeats.

		Beware that receiving a successful response does not necessarily mean that
		the heartbeat exists: https://docs.opsgenie.com/docs/heartbeat-api#ping-heartbeat-request

		Heartbeats to enable can be specified by a combination of selectors all of
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
	pingDocExamples = heredoc.Doc(`
		# ping all heartbeats with specified label 'managed-by' equal to 'foobricator'
		heartbeatctl ping --selector=managed-by=foobricator
		# ping all disabled heartbeats with label 'managed-by' equal to 'foobricator'
		heartbeatctl ping --selector="!enabled,managed-by=foobricator"
		# ping all heartbeats with alert priority equal to 'P3'
		heartbeatctl ping --field-selector=alertPriority=P3
		# ping all non-expired heartbeats with alert priority equal to 'P2' or 'P4'
		heartbeatctl ping -l "!expired,alertPriority in (P2, P4)"
		# ping expired heartbeats with alert priority equal to 'P3'
		heartbeatctl ping -l expired --field-selector=alertPriority=P3
		# ping all expired and active heartbeats
		heartbeatctl ping -l expired,enabled
		# ping heartbeats with exact names
		heartbeatctl ping foo foo-rab1 bar-oof2
		# ping heartbeats with names matching any of the regular expressions
		heartbeatctl ping "foo.*" ".*-oof[12]"
		# ping enabled heartbeats with alert priority equal to 'P3' but only those
		# with names matching any of the given regular expressions
		heartbeatctl ping -l enabled --field-selector=alertPriority=P3 "foo.*" ".*-oof[12]"
		# ping all heartbeats (note that an explicit selector matching everything
		# must be given)
		heartbeatctl ping ".*"
	`)
)

func init() {
	rootCmd.AddCommand(NewCmdPing())
}

func NewPingOptions() *pingCmdOptions {
	return &pingCmdOptions{
		selectorOptions: cmdutil.NewSelectorOptions(),
	}
}

func NewCmdPing() *cobra.Command {
	opts := NewPingOptions()

	cmd := &cobra.Command{
		Use:     "ping [NAME..]",
		Short:   "Ping heartbeats",
		Long:    pingDocLong,
		Example: pingDocExamples,
		Run: func(cmd *cobra.Command, args []string) {
			runPing(opts)
		},
	}

	opts.selectorOptions.WithCapturingArgsUsingValidator().AddFlags(cmd)

	return cmd
}

func runPing(opts *pingCmdOptions) {
	repo, err := client.New(nil)
	if err != nil {
		log.Fatalf("Failed to init OpsGenie client: %v\n", err)
	}
	c := ctl.NewCtl(repo)
	pings, err := c.Ping(opts.selectorOptions.ToConfig())
	if err != nil {
		log.Fatalf("Failed to ping heartbeats: %v\n", err)
	}
	for name, ping := range pings {
		fmt.Printf("heartbeat \"%s\": %s\n", name, ping.Message)
	}
}
