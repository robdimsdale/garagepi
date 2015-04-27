package gpio_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/garagepi"
	"github.com/robdimsdale/garagepi/gpio"
	oshelper_fakes "github.com/robdimsdale/garagepi/oshelper/fakes"
)

const (
	gpioPin        = uint(1)
	gpioExecutable = "gpio"
)

var (
	fakeOsHelper *oshelper_fakes.FakeOsHelper
	g            gpio.Gpio
)

var _ = Describe("Gpio", func() {
	BeforeEach(func() {
		fakeOsHelper = &oshelper_fakes.FakeOsHelper{}
		g = gpio.NewGpio(fakeOsHelper, gpioExecutable)
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
					"1",
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
})
