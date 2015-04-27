package homepage_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHomepage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Homepage Suite")
}
