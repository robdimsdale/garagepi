package garagepi_test

import (
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"

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

func tostr(u uint) string {
	return strconv.FormatUint(uint64(u), 10)
}

var _ = Describe("Garagepi", func() {

	webcamHost := "webcamHost"
	webcamPort := uint(12345)

	gpioDoorPin := uint(0)
	gpioLightPin := uint(8)
	gpioExecutable := "gpio"

	executorConfig := garagepi.ExecutorConfig{
		WebcamHost:     webcamHost,
		WebcamPort:     webcamPort,
		GpioDoorPin:    gpioDoorPin,
		GpioLightPin:   gpioLightPin,
		GpioExecutable: gpioExecutable,
	}

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
			fakeOsHelper,
			fakeFsHelper,
			fakeHttpHelper,
			executorConfig,
		)
	})

	Describe("Homepage Handling", func() {
		Context("When reading the homepage template is successful", func() {

			contents := "templateContents"
			BeforeEach(func() {
				t, err := template.New("template").Parse(contents)
				Expect(err).NotTo(HaveOccurred())
				fakeFsHelper.GetHomepageTemplateReturns(t, nil)
			})

			It("Should write the contents of the homepage template to the response writer", func() {
				executor.HomepageHandler(fakeResponseWriter, dummyRequest)
				Expect(fakeFsHelper.GetHomepageTemplateCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte(contents)))
			})
		})

		Context("When reading the homepage template fails with error", func() {
			BeforeEach(func() {
				fakeFsHelper.GetHomepageTemplateReturns(nil, errors.New("Failed to read contents"))
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
				if executable != gpioExecutable {
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

			Expect(args[0]).To(Equal([]string{garagepi.GpioWriteCommand, tostr(gpioDoorPin), garagepi.GpioHighState}))
			Expect(args[1]).To(Equal([]string{garagepi.GpioWriteCommand, tostr(gpioDoorPin), garagepi.GpioLowState}))
		}

		verifyGpioWriteHighThenNoFurtherGpioCalls := func() {
			gpioCalls := 0

			args := make([]string, 1)

			for i := 0; i < fakeOsHelper.ExecCallCount(); i++ {
				executable, curArgs := fakeOsHelper.ExecArgsForCall(i)
				if executable != gpioExecutable {
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
			Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, tostr(gpioDoorPin), garagepi.GpioHighState}))
		}

		Context("When executing "+gpioExecutable+" commands return sucessfully", func() {
			It("Should write "+garagepi.GpioHighState+" to gpio "+tostr(gpioDoorPin)+", sleep, and write "+garagepi.GpioLowState+" to gpio "+tostr(gpioDoorPin), func() {
				executor.ToggleDoorHandler(fakeResponseWriter, dummyRequest)
				Expect(fakeOsHelper.SleepArgsForCall(0)).To(Equal(garagepi.SleepTime))
				verifyGpioWriteHighFirstThenWriteLow()
			})

			It("Should return 'door toggled'", func() {
				executor.ToggleDoorHandler(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("door toggled")))
			})
		})

		Context("When executing the first "+gpioExecutable+" command returns with errors", func() {
			BeforeEach(func() {
				fakeOsHelper.ExecStub = func(executable string, _ ...string) (string, error) {
					if executable == gpioExecutable {
						return "", errors.New(garagepi.GpioWriteCommand + " " + gpioExecutable + "error")
					}
					return "", nil
				}
			})

			It("Should not sleep or execute further gpio commands", func() {
				executor.ToggleDoorHandler(fakeResponseWriter, dummyRequest)
				Expect(fakeOsHelper.SleepCallCount()).To(Equal(0))
				verifyGpioWriteHighThenNoFurtherGpioCalls()
			})

			It("Should return 'error - door not toggled'", func() {
				executor.ToggleDoorHandler(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("error - door not toggled")))
			})
		})
	})

	Describe("Light handling", func() {
		Describe("Reading state", func() {
			Context("When reading light state returns with error", func() {
				BeforeEach(func() {
					fakeOsHelper.ExecReturns("", errors.New(garagepi.GpioReadCommand+" "+gpioExecutable+"error"))
				})

				It("Should read from gpio "+tostr(gpioLightPin), func() {
					executor.GetLightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))
					executable, args := fakeOsHelper.ExecArgsForCall(0)
					Expect(executable).To(Equal("gpio"))
					Expect(args).To(Equal([]string{garagepi.GpioReadCommand, tostr(gpioLightPin)}))
				})
				It("Returns 'unknown'", func() {
					executor.GetLightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
					Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("error - light state: unknown")))
				})
			})

			Context("When reading light state contains leading/trailing whitespace", func() {
				BeforeEach(func() {
					fakeOsHelper.ExecReturns("\t0\n", nil)
				})
				It("Strips whitespace", func() {
					executor.GetLightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
					Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("light state: off")))
				})
			})

			Context("When reading light state returns 0", func() {
				BeforeEach(func() {
					fakeOsHelper.ExecReturns("0", nil)
				})
				It("Returns 'light state: off'", func() {
					executor.GetLightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
					Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("light state: off")))
				})
			})

			Context("When reading light state returns 1", func() {
				BeforeEach(func() {
					fakeOsHelper.ExecReturns("1", nil)
				})
				It("Returns 'light state: on'", func() {
					executor.GetLightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
					Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("light state: on")))
				})
			})

			Context("When reading light state returns unrecognized number", func() {
				BeforeEach(func() {
					fakeOsHelper.ExecReturns("2", nil)
				})
				It("Returns 'error - light state: unknown'", func() {
					executor.GetLightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
					Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("error - light state: unknown")))
				})
			})
		})

		Describe("Setting state", func() {
			Context("When attempting to set state without state information", func() {
				BeforeEach(func() {
					u, err := url.Parse("/?state")
					Expect(err).ShouldNot(HaveOccurred())
					dummyRequest.URL = u
				})
				It("Should write "+garagepi.GpioHighState+" to gpio "+tostr(gpioLightPin), func() {
					executor.SetLightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))
					executable, args := fakeOsHelper.ExecArgsForCall(0)
					Expect(executable).To(Equal("gpio"))
					Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, tostr(gpioLightPin), garagepi.GpioHighState}))
				})

				It("Should return 'light state: on'", func() {
					executor.SetLightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
					Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("light state: on")))
				})
			})

			Context("When attempting to set state with incorrect state information", func() {
				BeforeEach(func() {
					u, err := url.Parse("/?state=somefakevalue")
					Expect(err).ShouldNot(HaveOccurred())
					dummyRequest.URL = u
				})
				It("Should write "+garagepi.GpioHighState+" to gpio "+tostr(gpioLightPin), func() {
					executor.SetLightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))
					executable, args := fakeOsHelper.ExecArgsForCall(0)
					Expect(executable).To(Equal("gpio"))
					Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, tostr(gpioLightPin), garagepi.GpioHighState}))
				})

				It("Should return 'light state: on'", func() {
					executor.SetLightHandler(fakeResponseWriter, dummyRequest)
					Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
					Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("light state: on")))
				})
			})

			Describe("Turning light on", func() {
				BeforeEach(func() {
					u, err := url.Parse("/?state=on")
					Expect(err).ShouldNot(HaveOccurred())
					dummyRequest.URL = u
				})

				Context("When turning on light commands returns with error", func() {
					BeforeEach(func() {
						fakeOsHelper.ExecReturns("", errors.New(garagepi.GpioWriteCommand+" "+gpioExecutable+"error"))
					})
					It("Should write "+garagepi.GpioHighState+" to gpio "+tostr(gpioLightPin), func() {
						executor.SetLightHandler(fakeResponseWriter, dummyRequest)
						Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))
						executable, args := fakeOsHelper.ExecArgsForCall(0)
						Expect(executable).To(Equal(gpioExecutable))
						Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, tostr(gpioLightPin), garagepi.GpioHighState}))
					})

					It("Should return 'error - light state unchanged'", func() {
						executor.SetLightHandler(fakeResponseWriter, dummyRequest)
						Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
						Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("error - light state unchanged")))
					})
				})

				Context("When turning on light command returns sucessfully", func() {
					It("Should write "+garagepi.GpioHighState+" to gpio "+tostr(gpioLightPin), func() {
						executor.SetLightHandler(fakeResponseWriter, dummyRequest)
						Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))
						executable, args := fakeOsHelper.ExecArgsForCall(0)
						Expect(executable).To(Equal("gpio"))
						Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, tostr(gpioLightPin), garagepi.GpioHighState}))
					})

					It("Should return 'light state: on'", func() {
						executor.SetLightHandler(fakeResponseWriter, dummyRequest)
						Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
						Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("light state: on")))
					})
				})
			})
			Describe("Turning light off", func() {
				BeforeEach(func() {
					u, err := url.Parse("/?state=off")
					Expect(err).ShouldNot(HaveOccurred())
					dummyRequest.URL = u
				})

				Context("When turning off light command returns with error", func() {
					BeforeEach(func() {
						fakeOsHelper.ExecReturns("", errors.New(gpioExecutable+" "+garagepi.GpioWriteCommand+"error"))
					})
					It("Should write "+garagepi.GpioLowState+" to gpio "+tostr(gpioLightPin), func() {
						executor.SetLightHandler(fakeResponseWriter, dummyRequest)
						Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))
						executable, args := fakeOsHelper.ExecArgsForCall(0)
						Expect(executable).To(Equal(gpioExecutable))
						Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, tostr(gpioLightPin), garagepi.GpioLowState}))
					})

					It("Should return 'light state unchanged'", func() {
						executor.SetLightHandler(fakeResponseWriter, dummyRequest)
						Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
						Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("error - light state unchanged")))
					})
				})

				Context("When turning off light command return sucessfully", func() {
					It("Should write "+garagepi.GpioLowState+" to gpio "+tostr(gpioLightPin), func() {
						executor.SetLightHandler(fakeResponseWriter, dummyRequest)
						Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))
						executable, args := fakeOsHelper.ExecArgsForCall(0)
						Expect(executable).To(Equal("gpio"))
						Expect(args).To(Equal([]string{garagepi.GpioWriteCommand, tostr(gpioLightPin), garagepi.GpioLowState}))
					})

					It("Should return 'light state: off'", func() {
						executor.SetLightHandler(fakeResponseWriter, dummyRequest)
						Expect(fakeResponseWriter.WriteCallCount()).To(Equal(1))
						Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal([]byte("light state: off")))
					})
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
