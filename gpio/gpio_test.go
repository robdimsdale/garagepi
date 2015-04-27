package gpio_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/garagepi"
	"github.com/robdimsdale/garagepi/gpio"
	logger_fakes "github.com/robdimsdale/garagepi/logger/fakes"
	oshelper_fakes "github.com/robdimsdale/garagepi/oshelper/fakes"
)

const (
	gpioExecutable = "gpio"
	gpioHighState  = "1"
	gpioLowState   = "0"

	gpioPin         = uint(1)
	gpioPinAsString = "1"
)

var (
	fakeOsHelper *oshelper_fakes.FakeOsHelper
	fakeLogger   *logger_fakes.FakeLogger
	g            gpio.Gpio
)

var _ = Describe("Gpio", func() {
	BeforeEach(func() {
		fakeOsHelper = &oshelper_fakes.FakeOsHelper{}
		fakeLogger = &logger_fakes.FakeLogger{}

		g = gpio.NewGpio(fakeOsHelper, fakeLogger, gpioExecutable)
	})

	Describe("Read", func() {
		Context("when osHelper returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("exec error")
				fakeOsHelper.ExecReturns("", expectedErr)
			})

			It("forwards the error", func() {
				_, err := g.Read(gpioPin)
				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("when osHelper returns sucessfully", func() {
			var expectedOutput string

			BeforeEach(func() {
				expectedOutput = "exec output"
				fakeOsHelper.ExecReturns(expectedOutput, nil)
			})

			It("returns the output without error", func() {
				expectedArgs := []string{
					garagepi.GpioReadCommand,
					gpioPinAsString,
				}

				output, err := g.Read(gpioPin)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(expectedOutput))

				Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))

				executable, args := fakeOsHelper.ExecArgsForCall(0)
				Expect(executable).To(Equal(gpioExecutable))
				Expect(args).To(Equal(expectedArgs))
			})
		})
	})

	Describe("WriteLow", func() {
		Context("when osHelper returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("exec error")
				fakeOsHelper.ExecReturns("", expectedErr)
			})

			It("forwards the error", func() {
				err := g.WriteLow(gpioPin)
				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("when osHelper returns sucessfully", func() {
			var expectedOutput string

			BeforeEach(func() {
				expectedOutput = "exec output"
				fakeOsHelper.ExecReturns(expectedOutput, nil)
			})

			It("returns without error", func() {
				expectedArgs := []string{
					garagepi.GpioWriteCommand,
					gpioPinAsString,
					gpioLowState,
				}

				err := g.WriteLow(gpioPin)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))

				executable, args := fakeOsHelper.ExecArgsForCall(0)
				Expect(executable).To(Equal(gpioExecutable))
				Expect(args).To(Equal(expectedArgs))
			})
		})
	})

	Describe("WriteHigh", func() {
		Context("when osHelper returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("exec error")
				fakeOsHelper.ExecReturns("", expectedErr)
			})

			It("forwards the error", func() {
				err := g.WriteHigh(gpioPin)
				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("when osHelper returns sucessfully", func() {
			var expectedOutput string

			BeforeEach(func() {
				expectedOutput = "exec output"
				fakeOsHelper.ExecReturns(expectedOutput, nil)
			})

			It("returns without error", func() {
				expectedArgs := []string{
					garagepi.GpioWriteCommand,
					gpioPinAsString,
					gpioHighState,
				}

				err := g.WriteHigh(gpioPin)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeOsHelper.ExecCallCount()).To(Equal(1))

				executable, args := fakeOsHelper.ExecArgsForCall(0)
				Expect(executable).To(Equal(gpioExecutable))
				Expect(args).To(Equal(expectedArgs))
			})
		})
	})

})
