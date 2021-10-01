package cmdutil_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/giantswarm/heartbeatctl/pkg/cmdutil"
)

var _ = Describe("Selector", func() {
	var (
		cmd     *cobra.Command
		opts    *cmdutil.SelectorOptions
		execute func([]string) error
	)

	BeforeEach(func() {
		cmd = &cobra.Command{
			Use: "fake",
			Run: func(_ *cobra.Command, _ []string) {},
		}
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		execute = func(a []string) error {
			cmd.SetArgs(a)
			return cmd.Execute()
		}

		opts = cmdutil.NewSelectorOptions()
	})

	Context("with Cobra command", func() {
		JustBeforeEach(func() {
			opts.AddFlags(cmd)
		})

		When("handling of args is not requested", func() {
			It("registers label and field selector flags with the command", func() {
				By("checking declared flags")
				flags := cmd.Flags()
				Expect(flags.Lookup("selector")).To(SatisfyAll(
					Not(BeNil()),
					WithTransform(func(f *pflag.Flag) string {
						return f.Shorthand
					}, Equal("l")),
				))
				Expect(flags.Lookup("field-selector")).NotTo(BeNil())

				By("parsing some flags and arguments")
				Expect(execute([]string{
					"--selector=enabled", "--field-selector=alertPriority=P2", "foo", "bar.*",
				})).To(Succeed())
				cfg := opts.ToConfig()

				By("checking selector options are populated but positional args are not captured by default")
				Expect(cfg.LabelSelector).To(Equal("enabled"))
				Expect(cfg.FieldSelector).To(Equal("alertPriority=P2"))
				Expect(cfg.NameExpressions).To(BeEmpty())
			})

			It("allows passing in name expressions explicitly", func() {
				By("adding name expressions to options")
				opts.NameExpressions("foo", "bar.*")

				By("checking they can be found in resulting config")
				Expect(opts.ToConfig().NameExpressions).To(ConsistOf("foo", "bar.*"))
			})
		})

		When("handling of args is requested", func() {
			BeforeEach(func() {
				v := opts.WithCapturingArgsUsingValidator()
				Expect(v).To(BeIdenticalTo(opts))
			})

			It("registers an arg handler with the command's .Args validator", func() {
				By("checking args validator function")
				Expect(cmd.Args).NotTo(BeNil())

				By("parsing some flags and arguments")
				Expect(execute([]string{
					"--selector=enabled", "--field-selector=alertPriority=P2", "foo", "bar.*",
				})).To(Succeed())
				cfg := opts.ToConfig()

				By("checking positional args are captured")
				Expect(cfg.NameExpressions).To(ConsistOf("foo", "bar.*"))
			})
		})
	})
})
