package light_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	test_helpers_fakes "github.com/robdimsdale/garagepi/fakes"
	gpio_fakes "github.com/robdimsdale/garagepi/gpio/fakes"
	"github.com/robdimsdale/garagepi/light"
)

const (
	gpioLightPin = uint(1)
)

var (
	fakeLogger         lager.Logger
	fakeGpio           *gpio_fakes.FakeGpio
	fakeResponseWriter *test_helpers_fakes.FakeResponseWriter

	dummyRequest *http.Request
	lh           light.Handler
)

var _ = Describe("Light", func() {
	var expectedLightState light.LightState
	var expectedReturn []byte
	var err error

	BeforeEach(func() {
		expectedLightState = light.LightState{
			StateKnown: false,
			LightOn:    false,
		}

		fakeLogger = lagertest.NewTestLogger("light test")
		fakeGpio = new(gpio_fakes.FakeGpio)
		fakeResponseWriter = new(test_helpers_fakes.FakeResponseWriter)

		lh = light.NewHandler(
			fakeLogger,
			fakeGpio,
			gpioLightPin,
		)

		dummyRequest = new(http.Request)
	})

	Describe("Reading state", func() {
		Context("When reading light state returns with error", func() {
			BeforeEach(func() {
				fakeGpio.ReadReturns("", errors.New("gpio read error"))
				expectedLightState.StateKnown = false
				expectedReturn, err = json.Marshal(expectedLightState)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should read from light pin", func() {
				lh.HandleGet(fakeResponseWriter, dummyRequest)
				Expect(fakeGpio.ReadCallCount()).To(Equal(1))
				Expect(fakeGpio.ReadArgsForCall(0)).To(Equal(gpioLightPin))
			})

			It("Should return unknown light state", func() {
				lh.HandleGet(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(expectedReturn))
			})

			It("Should respond with HTTP status code 503", func() {
				lh.HandleGet(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteHeaderCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteHeaderArgsForCall(0)).To(Equal(http.StatusServiceUnavailable))
			})
		})

		Context("When reading light state contains leading/trailing whitespace", func() {
			BeforeEach(func() {
				fakeGpio.ReadReturns("\t0\n", nil)
				expectedLightState.StateKnown = true
				expectedLightState.LightOn = false
				expectedReturn, err = json.Marshal(expectedLightState)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should strip whitespace", func() {
				lh.HandleGet(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(expectedReturn))
			})
		})

		Context("When reading light state returns 0", func() {
			BeforeEach(func() {
				fakeGpio.ReadReturns("0", nil)
				expectedLightState.StateKnown = true
				expectedLightState.LightOn = false
				expectedReturn, err = json.Marshal(expectedLightState)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should return light state off", func() {
				lh.HandleGet(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(expectedReturn))
			})
		})

		Context("When reading light state returns 1", func() {
			BeforeEach(func() {
				fakeGpio.ReadReturns("1", nil)
				expectedLightState.StateKnown = true
				expectedLightState.LightOn = true
				expectedReturn, err = json.Marshal(expectedLightState)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should return light state on", func() {
				lh.HandleGet(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(expectedReturn))
			})
		})

		Context("When reading light state returns unrecognized number", func() {
			BeforeEach(func() {
				fakeGpio.ReadReturns("2", nil)
				expectedLightState.StateKnown = false
				expectedReturn, err = json.Marshal(expectedLightState)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should respond with HTTP status code 503", func() {
				lh.HandleGet(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteHeaderCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteHeaderArgsForCall(0)).To(Equal(http.StatusServiceUnavailable))
			})

			It("Should return unknown light state", func() {
				lh.HandleGet(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(expectedReturn))
			})
		})
	})

	Describe("Setting state", func() {
		Context("When attempting to set state without state information", func() {
			BeforeEach(func() {
				u, err := url.Parse("/?state")
				Expect(err).ShouldNot(HaveOccurred())
				dummyRequest.URL = u

				expectedLightState.StateKnown = true
				expectedLightState.LightOn = true
				expectedReturn, err = json.Marshal(expectedLightState)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should write high to light pin", func() {
				lh.HandleSet(fakeResponseWriter, dummyRequest)

				Expect(fakeGpio.WriteHighCallCount()).To(Equal(1))

				actualGpioPin := fakeGpio.WriteHighArgsForCall(0)
				Expect(actualGpioPin).To(Equal(gpioLightPin))
			})

			It("Should return light state on", func() {
				lh.HandleSet(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(expectedReturn))
			})
		})

		Context("When attempting to set state with incorrect state information", func() {
			BeforeEach(func() {
				u, err := url.Parse("/?state=somefakevalue")
				Expect(err).ShouldNot(HaveOccurred())
				dummyRequest.URL = u

				expectedLightState.StateKnown = true
				expectedLightState.LightOn = true
				expectedReturn, err = json.Marshal(expectedLightState)
				Expect(err).NotTo(HaveOccurred())
			})

			It("Should write high to light pin", func() {
				lh.HandleSet(fakeResponseWriter, dummyRequest)

				Expect(fakeGpio.WriteHighCallCount()).To(Equal(1))

				actualGpioPin := fakeGpio.WriteHighArgsForCall(0)
				Expect(actualGpioPin).To(Equal(gpioLightPin))
			})

			It("Should return light state on", func() {
				lh.HandleSet(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(expectedReturn))
			})
		})

		Describe("Turning light on", func() {
			BeforeEach(func() {
				u, err := url.Parse("/?state=on")
				Expect(err).ShouldNot(HaveOccurred())
				dummyRequest.URL = u

				expectedLightState.StateKnown = true
				expectedLightState.LightOn = true
				expectedReturn, err = json.Marshal(expectedLightState)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("When turning on light commands returns with error", func() {
				BeforeEach(func() {
					expectedError := errors.New(fmt.Sprintf("gpio write error"))
					fakeGpio.WriteHighReturns(expectedError)

					expectedLightState.StateKnown = false
					expectedLightState.LightOn = false
					expectedLightState.ErrorMsg = expectedError.Error()
					expectedReturn, err = json.Marshal(expectedLightState)
					Expect(err).NotTo(HaveOccurred())
				})

				It("Should write high to light pin", func() {
					lh.HandleSet(fakeResponseWriter, dummyRequest)

					Expect(fakeGpio.WriteHighCallCount()).To(Equal(1))

					actualGpioPin := fakeGpio.WriteHighArgsForCall(0)
					Expect(actualGpioPin).To(Equal(gpioLightPin))
				})

				It("Should return light state unknown", func() {
					lh.HandleSet(fakeResponseWriter, dummyRequest)
					Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
					Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(expectedReturn))
				})
			})

			Context("When turning on light command returns sucessfully", func() {
				It("Should write high to light pin", func() {
					lh.HandleSet(fakeResponseWriter, dummyRequest)

					Expect(fakeGpio.WriteHighCallCount()).To(Equal(1))

					actualGpioPin := fakeGpio.WriteHighArgsForCall(0)
					Expect(actualGpioPin).To(Equal(gpioLightPin))
				})

				It("Should return light state on", func() {
					lh.HandleSet(fakeResponseWriter, dummyRequest)
					Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
					Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(expectedReturn))
				})
			})
		})

		Describe("Turning light off", func() {
			BeforeEach(func() {
				u, err := url.Parse("/?state=off")
				Expect(err).ShouldNot(HaveOccurred())
				dummyRequest.URL = u

				expectedLightState.StateKnown = true
				expectedLightState.LightOn = false
				expectedReturn, err = json.Marshal(expectedLightState)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("When turning off light command returns with error", func() {
				BeforeEach(func() {
					expectedError := errors.New(fmt.Sprintf("gpio write error"))
					fakeGpio.WriteLowReturns(expectedError)

					expectedLightState.StateKnown = false
					expectedLightState.LightOn = false
					expectedLightState.ErrorMsg = expectedError.Error()
					expectedReturn, err = json.Marshal(expectedLightState)
					Expect(err).NotTo(HaveOccurred())
				})

				It("Should write low to light pin", func() {
					lh.HandleSet(fakeResponseWriter, dummyRequest)

					Expect(fakeGpio.WriteLowCallCount()).To(Equal(1))

					actualGpioPin := fakeGpio.WriteLowArgsForCall(0)
					Expect(actualGpioPin).To(Equal(gpioLightPin))
				})

				It("Should return light state unknown", func() {
					lh.HandleSet(fakeResponseWriter, dummyRequest)
					Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
					Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(expectedReturn))
				})
			})

			Context("When turning off light command return sucessfully", func() {
				It("Should write low to light pin", func() {
					lh.HandleSet(fakeResponseWriter, dummyRequest)

					Expect(fakeGpio.WriteLowCallCount()).To(Equal(1))

					actualGpioPin := fakeGpio.WriteLowArgsForCall(0)
					Expect(actualGpioPin).To(Equal(gpioLightPin))
				})

				It("Should return light state off", func() {
					lh.HandleSet(fakeResponseWriter, dummyRequest)
					Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
					Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(expectedReturn))
				})
			})
		})
	})
})
