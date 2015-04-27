package oshelper_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logger_fakes "github.com/robdimsdale/garagepi/logger/fakes"
	"github.com/robdimsdale/garagepi/oshelper"
)

var _ = Describe("OsHelper", func() {
	var osHelper oshelper.OsHelper

	BeforeEach(func() {
		fakeLogger := &logger_fakes.FakeLogger{}
		osHelper = oshelper.NewOsHelperImpl(fakeLogger)
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
