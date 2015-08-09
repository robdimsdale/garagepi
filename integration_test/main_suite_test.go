package main_test

import (
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
	port            uint
	garagepiBinPath string
)

var _ = BeforeSuite(func() {
	var err error
	garagepiBinPath, err = gexec.Build("github.com/robdimsdale/garagepi", "-race")
	Expect(err).ShouldNot(HaveOccurred())

	port = uint(59990 + GinkgoParallelNode())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
