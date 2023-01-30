package cmdutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCmdutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmdutil Suite")
}
