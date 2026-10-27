package vocabjson_test

import (
	"github.com/kinglet-dev/saymake/adapters/vocabjson"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/families/flat"
	"github.com/kinglet-dev/saymake/families/tray"
	"github.com/kinglet-dev/saymake/families/vessel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Default", func() {
	It("loads the built-in vocabulary, and it fits the built-in families", func() {
		// Arrange
		r, err := family.NewRegistry(vessel.New(), tray.New(), flat.New())
		Expect(err).NotTo(HaveOccurred())

		// Act
		v, err := vocabjson.Default()

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Validate(r)).To(Succeed())
		Expect(len(v.Concepts)).To(BeNumerically(">=", 15))
	})
})
