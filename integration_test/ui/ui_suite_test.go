package ui_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"

	"testing"
)

func TestGaragepiUI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Garagepi UI Suite")
}

var (
	httpPort        uint
	httpsPort       uint
	garagepiBinPath string
	agoutiDriver    *agouti.WebDriver
)

var _ = BeforeSuite(func() {
	// The default of 1 second is not always enough for the webserver to start
	// handling requests.
	SetDefaultEventuallyTimeout(10 * time.Second)

	var err error
	garagepiBinPath, err = gexec.Build("github.com/robdimsdale/garagepi", "-race")
	Expect(err).ShouldNot(HaveOccurred())

	httpPort = uint(59990 + 2*GinkgoParallelNode())
	httpsPort = uint(59991 + 2*GinkgoParallelNode())

	agoutiDriver = agouti.PhantomJS()
	Expect(agoutiDriver.Start()).To(Succeed())
})

var _ = AfterSuite(func() {
	Expect(agoutiDriver.Stop()).To(Succeed())
	gexec.CleanupBuildArtifacts()
})
