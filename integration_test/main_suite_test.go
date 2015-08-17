package main_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestGaragepi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GaragepiExecutable Suite")
}

var (
	httpPort        uint
	httpsPort       uint
	garagepiBinPath string
)

var _ = BeforeSuite(func() {
	// The default of 1 second is always enough for the webserver to start
	// handling requests.
	SetDefaultEventuallyTimeout(10 * time.Second)

	var err error
	garagepiBinPath, err = gexec.Build("github.com/robdimsdale/garagepi", "-race")
	Expect(err).ShouldNot(HaveOccurred())

	httpPort = uint(59990 + 2*GinkgoParallelNode())
	httpsPort = uint(59991 + 2*GinkgoParallelNode())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
