package vocab_test

import (
	"github.com/kinglet-dev/saymake/domain/param"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vocabulary.Resolve", func() {
	It("inherits the family and merges defaults down the chain, child last", func() {
		// Arrange
		v := sample()

		// Act
		r, err := v.Resolve("bud-vase")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Family).To(Equal("vessel"))
		Expect(r.Path).To(Equal([]string{"Vessel", "Vase", "Bud vase"}))
		Expect(r.Defaults).To(Equal(param.Values{"height": 120}))
	})

	It("keeps a parent's default the child doesn't set", func() {
		// Arrange
		v := sample()
		v.Concepts[1].Defaults = param.Values{"height": 200, "flutes": 6}

		// Act
		r, err := v.Resolve("bud-vase")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Defaults).To(Equal(param.Values{"height": 120, "flutes": 6}))
	})

	It("uses the nearest family on the chain", func() {
		// Arrange
		v := sample()

		// Act
		r, err := v.Resolve("tray")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Family).To(Equal("tray"))
		Expect(r.Path).To(Equal([]string{"Tray"}))
	})

	It("refuses an id that isn't in the vocabulary", func() {
		// Arrange
		v := sample()

		// Act
		_, err := v.Resolve("lamp")

		// Assert
		Expect(err).To(MatchError(`concept "lamp" is not in the vocabulary`))
	})

	It("finds whether one concept is an ancestor of another", func() {
		// Arrange
		v := sample()

		// Act
		yes, no, self := v.IsAncestor("vessel", "bud-vase"), v.IsAncestor("tray", "vase"), v.IsAncestor("vase", "vase")

		// Assert
		Expect([]bool{yes, no, self}).To(Equal([]bool{true, false, false}))
	})
})

var _ = Describe("Modifier.AppliesTo", func() {
	It("applies to every family when none is listed", func() {
		// Arrange
		m := sample().Modifiers[1]

		// Act
		ok := m.AppliesTo("tray")

		// Assert
		Expect(ok).To(BeTrue())
	})

	It("applies only to the listed families", func() {
		// Arrange
		m := sample().Modifiers[0]

		// Act
		yes, no := m.AppliesTo("vessel"), m.AppliesTo("tray")

		// Assert
		Expect([]bool{yes, no}).To(Equal([]bool{true, false}))
	})
})
