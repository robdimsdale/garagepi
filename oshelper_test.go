package garagepi_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/robdimsdale/garagepi"
	fakes "github.com/robdimsdale/garagepi/fakes"
)

var _ = Describe("OsHelper", func() {
	var osHelper garagepi.OsHelper

	BeforeEach(func() {
		fakeLogger := &fakes.FakeLogger{}
		osHelper = garagepi.NewOsHelperImpl(fakeLogger)
	})

	Describe("Exec", func() {
		It("Executes 'echo hello world' successfully", func() {
			resp, err := osHelper.Exec("echo", "hello", "world")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).To(Equal("hello world\n"))
		})
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
