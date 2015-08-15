package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/robdimsdale/garagepi/middleware"
	"github.com/robdimsdale/garagepi/middleware/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HttpsEnforcer", func() {
	const httpsPort = uint(34567)

	var (
		request           *http.Request
		writer            *httptest.ResponseRecorder
		fakeHandler       *fakes.FakeHandler
		wrappedMiddleware http.Handler
	)

	BeforeEach(func() {
		fakeHandler = &fakes.FakeHandler{}
		writer = httptest.NewRecorder()
		enforcer := middleware.NewHTTPSEnforcer(httpsPort)

		wrappedMiddleware = enforcer.Wrap(fakeHandler)
	})

	Context("when the URL is valid", func() {
		BeforeEach(func() {
			var err error
			request, err = http.NewRequest("GET", "http://localhost/foo/bar", nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not call next middleware", func() {
			wrappedMiddleware.ServeHTTP(writer, request)

			Expect(fakeHandler.ServeHTTPCallCount()).To(BeZero())
		})

		It("redirects to https", func() {
			wrappedMiddleware.ServeHTTP(writer, request)

			Expect(writer.Code).To(Equal(http.StatusFound))
			expectedURL := fmt.Sprintf("https://localhost:%d/foo/bar", httpsPort)
			Expect(writer.HeaderMap.Get("Location")).To(Equal(expectedURL))
		})
	})

	Context("when the URL is invalid", func() {
		BeforeEach(func() {
			var err error
			request, err = http.NewRequest("GET", "http://localhost/foo/bar", nil)
			Expect(err).NotTo(HaveOccurred())

			request.URL.Host = "%%%"
		})

		It("should respond with a 401", func() {
			wrappedMiddleware.ServeHTTP(writer, request)

			Expect(writer.Code).To(Equal(http.StatusBadRequest))
			Expect(writer.Body.String()).To(Equal("Bad Request"))
		})
	})
})
