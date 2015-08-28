package door_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDoor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Door Suite")
}
