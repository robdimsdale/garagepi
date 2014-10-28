package garagepi_test

import (
	"errors"
	"io"
	"net/http"

	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/garagepi"
	garagepi_fakes "github.com/robdimsdale/garagepi/fakes"
)

var _ = Describe("Garagepi", func() {
	var fakeLogger *garagepi_fakes.FakeLogger
	var fakeHttpHelper *garagepi_fakes.FakeHttpHelper
	var fakeOsHelper *garagepi_fakes.FakeOsHelper
	var fakeResponseWriter *garagepi_fakes.FakeResponseWriter
	var dummyRequest *http.Request

	webcamHost := "webcamHost"
	webcamPort := "webcamPort"

	var executor *garagepi.Executor
	BeforeEach(func() {
		fakeLogger = new(garagepi_fakes.FakeLogger)
		fakeHttpHelper = new(garagepi_fakes.FakeHttpHelper)
		fakeOsHelper = new(garagepi_fakes.FakeOsHelper)

		fakeResponseWriter = new(garagepi_fakes.FakeResponseWriter)
		dummyRequest = new(http.Request)

		executor = garagepi.NewExecutor(
			fakeLogger,
			fakeHttpHelper,
			fakeOsHelper,
			webcamHost,
			webcamPort,
		)
	})

	Describe("Homepage Handling", func() {
		Context("When reading the homepage template is successful", func() {
			contents := []byte("contents")

			BeforeEach(func() {
				fakeOsHelper.GetHomepageTemplateContentsReturns(contents, nil)
			})

			It("Should write the contents of the homepage template to the response writer", func() {
				executor.HomepageHandler(fakeResponseWriter, dummyRequest)
				Expect(fakeOsHelper.GetHomepageTemplateContentsCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(contents))
			})
		})

		Context("When reading the homepage template fails with error", func() {
			BeforeEach(func() {
				fakeOsHelper.GetHomepageTemplateContentsReturns(nil, errors.New("Failed to read contents"))
			})

			execution := func() {
				executor.HomepageHandler(fakeResponseWriter, dummyRequest)
			}

			It("Should panic", func() {
				Expect(execution).Should(Panic())
			})
		})
	})

	Describe("Webcam handling", func() {
		Context("When obtaining a webcam image is successful", func() {
			contents := []byte("webcamImage")
			BeforeEach(func() {
				dummyResponse := new(http.Response)
				dummyResponse.Body = nopCloser{bytes.NewReader(contents)}
				fakeHttpHelper.GetReturns(dummyResponse, nil)
			})

			It("Should write the contents of the response to the response writer", func() {
				executor.WebcamHandler(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(contents))
			})
		})

		Context("When obtaining a webcam image fails with error", func() {
			BeforeEach(func() {
				fakeHttpHelper.GetReturns(nil, errors.New("Failed to GET url"))
			})

			It("Should write nothing to the response writer and return", func() {
				executor.WebcamHandler(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(0))
			})
		})
	})

	Describe("Door-toggle handling", func() {
		verifyGpioWriteHighFirstThenWriteLow := func() {
			gpioCalls := 0

			args := make([][]string, 2)

			for i := 0; i < fakeOsHelper.ExecCallCount(); i++ {
				executable, curArgs := fakeOsHelper.ExecArgsForCall(i)
				if executable != garagepi.GpioExecutable {
					continue
				}
				if gpioCalls == 0 {
					gpioCalls++
					args[0] = curArgs
				} else if gpioCalls == 1 {
					gpioCalls++
					args[1] = curArgs
				} else {
					gpioCalls++
				}
			}
			Expect(gpioCalls).To(Equal(2))

			Expect(args[0]).To(Equal([]string{garagepi.GpioWriteCommand, garagepi.GpioPin, garagepi.GpioHighState}))
			Expect(args[1]).To(Equal([]string{garagepi.GpioWriteCommand, garagepi.GpioPin, garagepi.GpioLowState}))
		}

		verifyGpioWriteHighThenNoFurtherGpioCalls := func() {
			gpioCalls := 0

			args := make([]string, 1)

			for i := 0; i < fakeOsHelper.ExecCallCount(); i++ {
				executable, curArgs := fakeOsHelper.ExecArgsForCall(i)
				if executable != garagepi.GpioExecutable {
					continue
				}
				if gpioCalls == 0 {
					gpioCalls++
					args = curArgs
				} else {
					gpioCalls++
				}
			}
			Expect(gpioCalls).To(Equal(1))
			Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, garagepi.GpioPin, garagepi.GpioHighState}))
		}

		verifyRedirectToHomepage := func() {
			Expect(fakeHttpHelper.RedirectToHomepageCallCount()).To(Equal(1))
			w, r := fakeHttpHelper.RedirectToHomepageArgsForCall(0)
			Expect(w).To(Equal(fakeResponseWriter))
			Expect(r).To(Equal(dummyRequest))
		}

		Context("When executing "+garagepi.GpioWriteCommand+" commands return sucessfully", func() {
			It("Should write "+garagepi.GpioHighState+" to gpio "+garagepi.GpioPin+", sleep, and write "+garagepi.GpioLowState+" to gpio "+garagepi.GpioPin, func() {
				executor.ToggleDoorHandler(fakeResponseWriter, dummyRequest)
				Expect(fakeOsHelper.SleepArgsForCall(0)).To(Equal(garagepi.SleepTime))
				verifyGpioWriteHighFirstThenWriteLow()
			})

			It("Should redirect to homepage", func() {
				executor.ToggleDoorHandler(fakeResponseWriter, dummyRequest)
				verifyRedirectToHomepage()
			})
		})

		Context("When executing the first "+garagepi.GpioWriteCommand+" command returns with errors", func() {
			BeforeEach(func() {
				fakeOsHelper.ExecStub = func(executable string, _ ...string) (string, error) {
					if executable == garagepi.GpioExecutable {
						return "", errors.New(garagepi.GpioExecutable + "error")
					}
					return "", nil
				}
			})

			It("Should not sleep or execute further gpio commands", func() {
				executor.ToggleDoorHandler(fakeResponseWriter, dummyRequest)
				Expect(fakeOsHelper.SleepCallCount()).To(Equal(0))
				verifyGpioWriteHighThenNoFurtherGpioCalls()
			})

			It("Should redirect to homepage", func() {
				executor.ToggleDoorHandler(fakeResponseWriter, dummyRequest)
				verifyRedirectToHomepage()
			})
		})
	})
})

type nopCloser struct {
	io.Reader
}

func (n nopCloser) Close() error {
	return nil
}
