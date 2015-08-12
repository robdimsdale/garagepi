package oshelper_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"

	"github.com/robdimsdale/garagepi/oshelper"
)

var _ = Describe("OsHelper", func() {
	var osHelper oshelper.OsHelper

	BeforeEach(func() {
		fakeLogger := lagertest.NewTestLogger("oshelper test")
		osHelper = oshelper.NewOsHelperImpl(fakeLogger)
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
