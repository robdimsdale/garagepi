package homepage_test

import (
	"errors"
	"html/template"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	fshelper_fakes "github.com/robdimsdale/garagepi/fshelper/fakes"
	"github.com/robdimsdale/garagepi/homepage"
	httphelper_fakes "github.com/robdimsdale/garagepi/httphelper/fakes"
	light_fakes "github.com/robdimsdale/garagepi/light/fakes"
	logger_fakes "github.com/robdimsdale/garagepi/logger/fakes"
	test_helpers_fakes "github.com/robdimsdale/garagepi/test_helpers/fakes"
)

var (
	fakeHttpHelper     *httphelper_fakes.FakeHttpHelper
	fakeLogger         *logger_fakes.FakeLogger
	fakeLightHandler   *light_fakes.FakeHandler
	fakeFsHelper       *fshelper_fakes.FakeFsHelper
	fakeResponseWriter *test_helpers_fakes.FakeResponseWriter

	dummyRequest *http.Request
	hh           homepage.Handler
)

var _ = Describe("Homepage", func() {

	BeforeEach(func() {
		fakeLogger = new(logger_fakes.FakeLogger)
		fakeLightHandler = new(light_fakes.FakeHandler)
		fakeFsHelper = new(fshelper_fakes.FakeFsHelper)
		fakeHttpHelper = new(httphelper_fakes.FakeHttpHelper)
		fakeResponseWriter = new(test_helpers_fakes.FakeResponseWriter)

		hh = homepage.NewHandler(
			fakeLogger,
			fakeHttpHelper,
			fakeFsHelper,
			fakeLightHandler,
		)

		dummyRequest = new(http.Request)
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
				hh.Handle(fakeResponseWriter, dummyRequest)
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
				hh.Handle(fakeResponseWriter, dummyRequest)
			}

			It("Should panic", func() {
				Expect(execution).Should(Panic())
			})
		})
	})
})
