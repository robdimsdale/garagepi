package light_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLight(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Light Suite")
}
