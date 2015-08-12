package door_test

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/robdimsdale/garagepi/door"
	test_helpers_fakes "github.com/robdimsdale/garagepi/fakes"
	gpio_fakes "github.com/robdimsdale/garagepi/gpio/fakes"
	httphelper_fakes "github.com/robdimsdale/garagepi/httphelper/fakes"
	oshelper_fakes "github.com/robdimsdale/garagepi/oshelper/fakes"
)

const (
	gpioDoorPin = uint(1)
)

var (
	fakeHttpHelper     *httphelper_fakes.FakeHttpHelper
	fakeOsHelper       *oshelper_fakes.FakeOsHelper
	fakeLogger         lager.Logger
	fakeGpio           *gpio_fakes.FakeGpio
	fakeResponseWriter *test_helpers_fakes.FakeResponseWriter

	dummyRequest *http.Request
	dh           door.Handler
)

var _ = Describe("Door", func() {
	BeforeEach(func() {
		fakeLogger = lagertest.NewTestLogger("Door test")
		fakeOsHelper = new(oshelper_fakes.FakeOsHelper)
		fakeGpio = new(gpio_fakes.FakeGpio)
		fakeHttpHelper = new(httphelper_fakes.FakeHttpHelper)
		fakeResponseWriter = new(test_helpers_fakes.FakeResponseWriter)

		dh = door.NewHandler(
			fakeLogger,
			fakeHttpHelper,
			fakeOsHelper,
			fakeGpio,
			gpioDoorPin)

		dummyRequest = new(http.Request)
	})

	Context("When toggling and sleeping return sucessfully", func() {
		It("Should write high to door pin, sleep, and write low to door pin", func() {
			dh.HandleToggle(fakeResponseWriter, dummyRequest)
			Expect(fakeOsHelper.SleepArgsForCall(0)).To(Equal(door.SleepTime))

			Expect(fakeGpio.WriteHighCallCount()).To(Equal(1))
			Expect(fakeGpio.WriteLowCallCount()).To(Equal(1))

			actualHighPin := fakeGpio.WriteHighArgsForCall(0)
			Expect(actualHighPin).To(Equal(gpioDoorPin))

			actualLowPin := fakeGpio.WriteLowArgsForCall(0)
			Expect(actualLowPin).To(Equal(gpioDoorPin))
		})

		It("Should return 'door toggled'", func() {
			dh.HandleToggle(fakeResponseWriter, dummyRequest)
			Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
			Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("door toggled")))
		})
	})

	Context("When writing high returns with errors", func() {
		BeforeEach(func() {
			fakeGpio.WriteHighReturns(errors.New("gpio error"))
		})

		It("Should not sleep or execute further gpio commands", func() {
			dh.HandleToggle(fakeResponseWriter, dummyRequest)
			Expect(fakeOsHelper.SleepCallCount()).To(Equal(0))

			Expect(fakeGpio.WriteHighCallCount()).To(Equal(1))

			actualHighPin := fakeGpio.WriteHighArgsForCall(0)
			Expect(actualHighPin).To(Equal(gpioDoorPin))
		})

		It("Should return 'error - door not toggled'", func() {
			dh.HandleToggle(fakeResponseWriter, dummyRequest)
			Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
			Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("error - door not toggled")))
		})
	})
})
