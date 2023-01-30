package ctl_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCtl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ctl Suite")
}
