package smoke_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSmokeGaragepi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Garagepi Smoke Tests Suite")
}

var (
	httpPort        uint
	httpsPort       uint
	garagepiBinPath string
)

var _ = BeforeSuite(func() {
	// The default of 1 second is not always enough for the webserver to start
	// handling requests.
	SetDefaultEventuallyTimeout(10 * time.Second)

	garagepiBinPath = os.Getenv("GARAGEPI_BIN_PATH")
	if garagepiBinPath == "" {
		fmt.Printf("garagepiBinPath not set - using 'garagepi'")
		garagepiBinPath = "garagepi"
	}

	httpPort = uint(59990 + 2*GinkgoParallelNode())
	httpsPort = uint(59991 + 2*GinkgoParallelNode())
})
