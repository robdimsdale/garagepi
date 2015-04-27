package gpio_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGpio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gpio Suite")
}
