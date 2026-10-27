package vocabjson_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVocabJSON(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vocabulary JSON Suite")
}
