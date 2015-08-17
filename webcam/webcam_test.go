package webcam_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	test_helpers_fakes "github.com/robdimsdale/garagepi/fakes"
	httphelper_fakes "github.com/robdimsdale/garagepi/httphelper/fakes"
	"github.com/robdimsdale/garagepi/webcam"
)

const (
	webcamHost = "webcam-host"
	webcamPort = uint(12345)
)

var (
	fakeHTTPHelper     *httphelper_fakes.FakeHTTPHelper
	fakeLogger         lager.Logger
	fakeResponseWriter *test_helpers_fakes.FakeResponseWriter

	dummyRequest *http.Request
	w            webcam.Handler
)

var _ = Describe("Webcam", func() {
	BeforeEach(func() {
		fakeLogger = lagertest.NewTestLogger("webcam test")
		fakeHTTPHelper = new(httphelper_fakes.FakeHTTPHelper)
		fakeResponseWriter = new(test_helpers_fakes.FakeResponseWriter)

		w = webcam.NewHandler(
			fakeLogger,
			fakeHTTPHelper,
			webcamHost,
			webcamPort)

		dummyRequest = new(http.Request)
	})

	Context("When obtaining a webcam image is successful", func() {
		contents := []byte("webcamImage")
		BeforeEach(func() {
			dummyResponse := new(http.Response)
			dummyResponse.Body = ioutil.NopCloser(bytes.NewReader(contents))
			fakeHTTPHelper.GetReturns(dummyResponse, nil)
		})

		It("Should write the contents of the response to the response writer", func() {
			w.Handle(fakeResponseWriter, dummyRequest)
			Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(contents))
		})
	})

	Context("When obtaining a webcam image fails with error", func() {
		BeforeEach(func() {
			fakeHTTPHelper.GetReturns(nil, errors.New("Failed to GET url"))
		})

		It("Should write nothing to the response writer and return", func() {
			w.Handle(fakeResponseWriter, dummyRequest)
			Expect(fakeResponseWriter.WriteCallCount()).To(Equal(0))
		})

		It("Should respond with HTTP status code 503", func() {
			w.Handle(fakeResponseWriter, dummyRequest)
			Expect(fakeResponseWriter.WriteHeaderCallCount()).To(Equal(1))
			Expect(fakeResponseWriter.WriteHeaderArgsForCall(0)).To(Equal(http.StatusServiceUnavailable))
		})
	})

	Context("When reading the webcam image fails with error", func() {
		BeforeEach(func() {
			dummyResponse := new(http.Response)
			dummyResponse.Body = errCloser{bytes.NewReader([]byte{})}
			fakeHTTPHelper.GetReturns(dummyResponse, nil)
		})

		It("Should write nothing to the response writer and return", func() {
			w.Handle(fakeResponseWriter, dummyRequest)
			Expect(fakeResponseWriter.WriteCallCount()).To(Equal(0))
		})

		It("Should respond with HTTP status code 503", func() {
			w.Handle(fakeResponseWriter, dummyRequest)
			Expect(fakeResponseWriter.WriteHeaderCallCount()).To(Equal(1))
			Expect(fakeResponseWriter.WriteHeaderArgsForCall(0)).To(Equal(http.StatusServiceUnavailable))
		})
	})
})

type errCloser struct {
	io.Reader
}

func (e errCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("ReadError")
}

func (e errCloser) Close() error {
	return nil
}
