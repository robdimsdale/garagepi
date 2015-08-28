package webcam_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	test_helpers_fakes "github.com/robdimsdale/garagepi/fakes"
	"github.com/robdimsdale/garagepi/web/webcam"
)

var (
	fakeLogger         lager.Logger
	fakeResponseWriter *test_helpers_fakes.FakeResponseWriter

	dummyRequest *http.Request
	w            webcam.Handler
)

var _ = Describe("Webcam", func() {
	var server *ghttp.Server

	BeforeEach(func() {
		server = ghttp.NewServer()
		webcamURL := server.URL()

		fakeLogger = lagertest.NewTestLogger("webcam test")
		fakeResponseWriter = new(test_helpers_fakes.FakeResponseWriter)

		w = webcam.NewHandler(
			fakeLogger,
			webcamURL,
		)

		dummyRequest = new(http.Request)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("obtaining an image from the upstream server", func() {
		It("should make a request to fetch the image", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/"),
				),
			)

			w.Handle(fakeResponseWriter, dummyRequest)
			Ω(server.ReceivedRequests()).Should(HaveLen(1))
		})

		Context("When obtaining a webcam image is successful", func() {
			contents := []byte("webcamImage")

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/"),
						ghttp.RespondWith(http.StatusOK, contents),
					),
				)
			})

			BeforeEach(func() {
				dummyResponse := new(http.Response)
				dummyResponse.Body = ioutil.NopCloser(bytes.NewReader(contents))
			})

			It("Should write the contents of the response to the response writer", func() {
				w.Handle(fakeResponseWriter, dummyRequest)
				Expect(fakeResponseWriter.WriteArgsForCall(0)).To(Equal(contents))
			})
		})

		Context("When obtaining a webcam image fails with error", func() {
			BeforeEach(func() {
				w = webcam.NewHandler(
					fakeLogger,
					"not-a-val!d-url",
				)
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

		Context("When a status code other than 200 was returned", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/"),
						ghttp.RespondWith(http.StatusNotFound, nil),
					),
				)
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
