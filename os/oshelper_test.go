package os_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"

	"github.com/robdimsdale/garagepi/os"
)

var _ = Describe("OsHelper", func() {
	var osHelper os.OSHelper

	BeforeEach(func() {
		fakeLogger := lagertest.NewTestLogger("os test")
		osHelper = os.NewOSHelper(fakeLogger)
	})

	Describe("Sleep", func() {
		It("Sleeps for at least the time provided", func() {
			start := time.Now()
			osHelper.Sleep(1 * time.Second)
			end := time.Now()
			Expect((end.Sub(start)).Seconds()).Should(BeNumerically(">", 1))
		})
	})
})
