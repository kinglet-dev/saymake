package threemf_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestThreeMF(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "3MF Suite")
}
