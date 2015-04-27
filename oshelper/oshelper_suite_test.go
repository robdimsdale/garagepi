package oshelper_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOsHelper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OsHelper Suite")
}
