package svg_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSVG(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SVG Suite")
}
