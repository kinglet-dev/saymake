package family_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFamily(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Family Suite")
}
