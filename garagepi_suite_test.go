package garagepi_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGaragepi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Garagepi Suite")
}
