package garagepi_test

import (
	"errors"
	"io"
	"net/http"
	"net/url"

	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/garagepi"
	garagepi_fakes "github.com/robdimsdale/garagepi/fakes"
)

var (
	fakeLogger         *garagepi_fakes.FakeLogger
	fakeHttpHelper     *garagepi_fakes.FakeHttpHelper
	fakeOsHelper       *garagepi_fakes.FakeOsHelper
	fakeFsHelper       *garagepi_fakes.FakeFsHelper
	fakeResponseWriter *garagepi_fakes.FakeResponseWriter
	dummyRequest       *http.Request
)

func verifyRedirectToHomepage() {
	Expect(fakeHttpHelper.RedirectToHomepageCallCount()).To(Equal(1))
	w, r := fakeHttpHelper.RedirectToHomepageArgsForCall(0)
	Expect(w).To(Equal(fakeResponseWriter))
	Expect(r).To(Equal(dummyRequest))
}

var _ = Describe("Garagepi", func() {

	webcamHost := "webcamHost"
	webcamPort := uint(12345)

	var executor *garagepi.Executor
	BeforeEach(func() {
		fakeLogger = new(garagepi_fakes.FakeLogger)
		fakeHttpHelper = new(garagepi_fakes.FakeHttpHelper)
		fakeOsHelper = new(garagepi_fakes.FakeOsHelper)
		fakeFsHelper = new(garagepi_fakes.FakeFsHelper)

		fakeResponseWriter = new(garagepi_fakes.FakeResponseWriter)
		dummyRequest = new(http.Request)

		executor = garagepi.NewExecutor(
			fakeLogger,
			fakeHttpHelper,
			fakeOsHelper,
			fakeFsHelper,
			webcamHost,
			webcamPort,
		)
	})

	Describe("Homepage Handling", func() {
		Context("When reading the homepage template is successful", func() {
			contents := []byte("contents")

			BeforeEach(func() {
				fakeFsHelper.GetHomepageTemplateContentsReturns(contents, nil)
			})

			It("Should write the contents of the homepage template to the response writer", func() {
				executor.HomepageHandler(fakeResponseWriter, dummyRequest)
				Expect(fakeFsHelper.GetHomepageTemplateContentsCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(contents))
			})
		})

		Context("When reading the homepage template fails with error", func() {
			BeforeEach(func() {
				fakeFsHelper.GetHomepageTemplateContentsReturns(nil, errors.New("Failed to read contents"))
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

		Context("When reading the webcam image fails with error", func() {
			BeforeEach(func() {
				dummyResponse := new(http.Response)
				dummyResponse.Body = errCloser{bytes.NewReader([]byte{})}
				fakeHttpHelper.GetReturns(dummyResponse, nil)
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

	Describe("Light-toggle handling", func() {
		Describe("Turning light on", func() {
			BeforeEach(func() {
				u, err := url.Parse("/?light=on")
				Expect(err).ShouldNot(HaveOccurred())
				dummyRequest.URL = u
			})

			Context("When turning on light commands returns with error", func() {
				BeforeEach(func() {

					fakeOsHelper.ExecStub = func(executable string, _ ...string) (string, error) {
						if executable == garagepi.GpioExecutable {
							return "", errors.New(garagepi.GpioExecutable + "error")
						}
						return "", nil
					}
				})
				It("Should write "+garagepi.GpioHighState+" to gpio "+garagepi.GpioLightPin, func() {
					executor.LightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))
					executable, args := fakeOsHelper.ExecArgsForCall(0)
					Expect(executable).To(Equal(garagepi.GpioExecutable))
					Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, garagepi.GpioLightPin, garagepi.GpioHighState}))
				})

				It("Should redirect to homepage", func() {
					executor.LightHandler(fakeResponseWriter, dummyRequest)
					verifyRedirectToHomepage()
				})
			})

			Context("When turning on light commands return sucessfully", func() {
				It("Should write "+garagepi.GpioHighState+" to gpio "+garagepi.GpioLightPin, func() {
					executor.LightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))
					executable, args := fakeOsHelper.ExecArgsForCall(0)
					Expect(executable).To(Equal("gpio"))
					Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, garagepi.GpioLightPin, garagepi.GpioHighState}))
				})

				It("Should redirect to homepage", func() {
					executor.LightHandler(fakeResponseWriter, dummyRequest)
					verifyRedirectToHomepage()
				})
			})
		})
		Describe("Turning light off", func() {
			BeforeEach(func() {
				u, err := url.Parse("/?state=off")
				Expect(err).ShouldNot(HaveOccurred())
				dummyRequest.URL = u
			})

			Context("When turning on light command returns with error", func() {
				BeforeEach(func() {

					fakeOsHelper.ExecStub = func(executable string, _ ...string) (string, error) {
						if executable == garagepi.GpioExecutable {
							return "", errors.New(garagepi.GpioExecutable + "error")
						}
						return "", nil
					}
				})
				It("Should write "+garagepi.GpioLowState+" to gpio "+garagepi.GpioLightPin, func() {
					executor.LightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))
					executable, args := fakeOsHelper.ExecArgsForCall(0)
					Expect(executable).To(Equal(garagepi.GpioExecutable))
					Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, garagepi.GpioLightPin, garagepi.GpioLowState}))
				})

				It("Should redirect to homepage", func() {
					executor.LightHandler(fakeResponseWriter, dummyRequest)
					verifyRedirectToHomepage()
				})
			})

			Context("When turning on light commands return sucessfully", func() {
				It("Should write "+garagepi.GpioLowState+" to gpio "+garagepi.GpioLightPin, func() {
					executor.LightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))
					executable, args := fakeOsHelper.ExecArgsForCall(0)
					Expect(executable).To(Equal("gpio"))
					Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, garagepi.GpioLightPin, garagepi.GpioLowState}))
				})

				It("Should redirect to homepage", func() {
					executor.LightHandler(fakeResponseWriter, dummyRequest)
					verifyRedirectToHomepage()
				})
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

type errCloser struct {
	io.Reader
}

func (e errCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("ReadError")
}

func (e errCloser) Close() error {
	return nil
}
