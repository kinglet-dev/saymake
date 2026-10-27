package stl_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSTL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "STL Suite")
}
