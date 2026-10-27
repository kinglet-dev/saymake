package meshcheck_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMeshcheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Meshcheck Suite")
}
