package homepage_test

import (
	"errors"
	"html/template"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	test_helpers_fakes "github.com/robdimsdale/garagepi/fakes"
	filesystem_fakes "github.com/robdimsdale/garagepi/filesystem/fakes"
	"github.com/robdimsdale/garagepi/homepage"
	httphelper_fakes "github.com/robdimsdale/garagepi/httphelper/fakes"
	light_fakes "github.com/robdimsdale/garagepi/light/fakes"
)

var (
	fakeHTTPHelper     *httphelper_fakes.FakeHTTPHelper
	fakeLogger         lager.Logger
	fakeLightHandler   *light_fakes.FakeHandler
	fakeFsHelper       *filesystem_fakes.FakeFileSystemHelper
	fakeResponseWriter *test_helpers_fakes.FakeResponseWriter

	dummyRequest *http.Request
	hh           homepage.Handler
)

var _ = Describe("Homepage", func() {

	BeforeEach(func() {
		fakeLogger = lagertest.NewTestLogger("homepage handle test")
		fakeLightHandler = new(light_fakes.FakeHandler)
		fakeFsHelper = new(filesystem_fakes.FakeFileSystemHelper)
		fakeHTTPHelper = new(httphelper_fakes.FakeHTTPHelper)
		fakeResponseWriter = new(test_helpers_fakes.FakeResponseWriter)

		hh = homepage.NewHandler(
			fakeLogger,
			fakeHTTPHelper,
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
