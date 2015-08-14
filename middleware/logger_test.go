package middleware_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/robdimsdale/garagepi/middleware"
	"github.com/robdimsdale/garagepi/middleware/fakes"
)

var _ = Describe("Logger", func() {

	var dummyRequest *http.Request
	var err error

	var fakeResponseWriter http.ResponseWriter
	var fakeHandler *fakes.FakeHandler
	var fakeLogger *fakes.FakeLogger

	BeforeEach(func() {
		dummyRequest, err = http.NewRequest("GET", "/some-url", nil)
		Expect(err).NotTo(HaveOccurred())
		dummyRequest.Header.Add("Authorization", "some auth")

		fakeResponseWriter = &fakes.FakeResponseWriter{}
		fakeHandler = &fakes.FakeHandler{}
		fakeLogger = &fakes.FakeLogger{}
	})

	It("should not log credentials", func() {
		loggerMiddleware := middleware.NewLogger(fakeLogger)
		loggerHandler := loggerMiddleware.Wrap(fakeHandler)

		loggerHandler.ServeHTTP(fakeResponseWriter, dummyRequest)

		Expect(fakeLogger.DebugCallCount()).To(Equal(1))
		_, arg1 := fakeLogger.DebugArgsForCall(0)
		loggedRequest := arg1[0]["request"].(middleware.LoggableHTTPRequest)
		Expect(loggedRequest.Header.Get("Authorization")).To(Equal(""))
	})

	It("should call next handler", func() {
		loggerMiddleware := middleware.NewLogger(fakeLogger)
		loggerHandler := loggerMiddleware.Wrap(fakeHandler)

		loggerHandler.ServeHTTP(fakeResponseWriter, dummyRequest)

		Expect(fakeHandler.ServeHTTPCallCount()).To(Equal(1))
		arg0, arg1 := fakeHandler.ServeHTTPArgsForCall(0)
		Expect(arg0).ToNot(BeNil())
		Expect(arg1).To(Equal(dummyRequest))
	})
})
