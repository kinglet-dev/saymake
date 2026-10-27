package match_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMatch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Match Suite")
}
