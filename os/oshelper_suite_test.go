package os_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOSHelper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OS Suite")
}
