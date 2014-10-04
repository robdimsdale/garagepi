package garagepi_test

import (
	"errors"
	"io"
	"net/http"

	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/robdimsdale/garage-pi/garagepi"
	garagepi_fakes "github.com/robdimsdale/garage-pi/garagepi/fakes"
	httphelper_fakes "github.com/robdimsdale/garage-pi/httphelper/fakes"
	logger_fakes "github.com/robdimsdale/garage-pi/logger/fakes"
	oshelper_fakes "github.com/robdimsdale/garage-pi/oshelper/fakes"
)

var _ = Describe("Garagepi", func() {
	var fakeLogger *logger_fakes.FakeLogger
	var fakeHttpHelper *httphelper_fakes.FakeHttpHelper
	var fakeOsHelper *oshelper_fakes.FakeOsHelper
	var fakeResponseWriter *garagepi_fakes.FakeResponseWriter
	var dummyRequest *http.Request

	webcamHost := "webcamHost"
	webcamPort := "webcamPort"

	var executor *Executor
	BeforeEach(func() {
		fakeLogger = new(logger_fakes.FakeLogger)
		fakeHttpHelper = new(httphelper_fakes.FakeHttpHelper)
		fakeOsHelper = new(oshelper_fakes.FakeOsHelper)

		fakeResponseWriter = new(garagepi_fakes.FakeResponseWriter)
		dummyRequest = new(http.Request)

		executor = NewExecutor(
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
})

type nopCloser struct {
	io.Reader
}

func (n nopCloser) Close() error {
	return nil
}
