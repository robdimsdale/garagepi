package webcam_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestWebcam(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webcam Suite")
}
