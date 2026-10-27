package interpret_test

import (
	"github.com/kinglet-dev/saymake/app/interpret"
	"github.com/kinglet-dev/saymake/domain/vocab"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interpreter.Read: edges", func() {
	It("suggests the only object when there is one", func() {
		// Arrange
		v := vocabulary()
		v.Concepts = []vocab.Concept{{ID: "tray", Label: "Tray", Family: "tray", Words: []string{"tray"}}}
		v.Sizes = nil
		in, err := interpret.New(v, registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		_, err = in.Read("blorp")

		// Assert
		Expect(err).To(MatchError("I couldn't tell what to make. Name an object, such as tray."))
	})

	It("writes \"an\" before an object that starts with a vowel", func() {
		// Arrange
		v := vocabulary()
		v.Concepts = append(v.Concepts, vocab.Concept{ID: "ewer", Label: "Ewer", Parent: "vessel", Words: []string{"ewer"}})
		in, err := interpret.New(v, registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		got, err := in.Read("tray ewer")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Notes).To(ContainElement(`"tray" also names an object; making an Ewer, the last one named.`))
	})

	It("reads a number before an unknown word as a number without a unit", func() {
		// Arrange
		in, err := interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		got, err := in.Read("vase 3 zorgles")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Notes).To(ContainElement(`"3" has no unit, so it was left out. Add mm, cm or in after it.`))
		Expect(got.Unknown).To(Equal([]string{"zorgles"}))
	})

	It("reads a number before an object word as a number without a unit", func() {
		// Arrange: bowl is concept 3 and the flute count is modifier 3, so
		// mixing up the two lists would read "2 bowls" as a count.
		in, err := interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		got, err := in.Read("2 bowls")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Concept).To(Equal("bowl"))
		Expect(got.Values["flutes"]).To(BeNumerically("~", 0, tol))
		Expect(got.Notes).To(ContainElement(`"2" has no unit, so it was left out. Add mm, cm or in after it.`))
	})

	It("doesn't guess at a count word given without a number", func() {
		// Arrange
		in, err := interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		got, err := in.Read("vase with grooves")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Unknown).To(Equal([]string{"grooves"}))
		Expect(got.Notes).To(BeEmpty())
	})

	It("still reads a count word with a typo", func() {
		// Arrange
		in, err := interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		got, err := in.Read("vase with 8 flutse")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["flutes"]).To(BeNumerically("~", 8, tol))
	})

	It("reads a number before a feature without a count as a number without a unit", func() {
		// Arrange
		in, err := interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		got, err := in.Read("2 fluted vase")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["flutes"]).To(BeNumerically("~", 12, tol))
		Expect(got.Notes).To(ContainElement(`"2" has no unit, so it was left out. Add mm, cm or in after it.`))
	})
})
