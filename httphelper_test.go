package garagepi_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/robdimsdale/garagepi"
)

var (
	loggingOn = false
)

var _ = Describe("HttpHelper", func() {
	var httpHelper garagepi.HttpHelper
	BeforeEach(func() {
		httpHelper = garagepi.NewHttpHelperImpl()
	})
	Describe("Get", func() {
		It("Gets successfully", func() {
			resp, err := httpHelper.Get("http://google.com")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Body).NotTo(BeNil())

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(body)).Should(BeNumerically(">", 0))
		})
	})
})
