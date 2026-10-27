package interpret_test

import (
	"github.com/kinglet-dev/saymake/app/interpret"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interpreter.Read: sizes", func() {
	var in *interpret.Interpreter

	BeforeEach(func() {
		// Arrange (shared)
		var err error
		in, err = interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())
	})

	It("gives a bare size to the family's main dimension", func() {
		// Arrange
		d := "vase 20 cm"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["height"]).To(BeNumerically("~", 200, tol))
		Expect(got.Uses).To(ContainElement(interpret.Use{Typed: "20 cm", Meaning: "Height 200 mm", Kind: interpret.Size}))
	})

	It("gives a size to the dimension named after it", func() {
		// Arrange
		d := "a bowl 8 in wide"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["diameter"]).To(BeNumerically("~", 203.2, tol))
		Expect(got.Uses).To(ContainElement(interpret.Use{Typed: "8 in wide", Meaning: "Diameter 203.2 mm", Kind: interpret.Size}))
	})

	It("gives a size to the dimension named before it, past small words", func() {
		// Arrange
		d := "vase with a height of 30cm"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["height"]).To(BeNumerically("~", 300, tol))
		Expect(got.Uses).To(ContainElement(interpret.Use{Typed: "height of 30 cm", Meaning: "Height 300 mm", Kind: interpret.Size}))
	})

	It("doesn't take an object word for a size word it is one typo from", func() {
		// Arrange: "vase" is one edit from the size word "base"
		d := "fluted vase 20 cm"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Concept).To(Equal("vase"))
		Expect(got.Values["height"]).To(BeNumerically("~", 200, tol))
	})

	It("doesn't take an object word after the size for a size word either", func() {
		// Arrange
		d := "20 cm vase"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Concept).To(Equal("vase"))
	})

	It("reads a word that is both a size word and a feature as the size, next to a size", func() {
		// Arrange: "wide" is a size word and also a feature (+30 mm)
		d := "bowl 10 cm wide"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["diameter"]).To(BeNumerically("~", 100, tol))
		Expect(got.Values["height"]).To(BeNumerically("~", 80, tol))
	})

	It("reads a size word at the very start", func() {
		// Arrange
		d := "height 30 cm vase"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["height"]).To(BeNumerically("~", 300, tol))
		Expect(got.Uses).To(ContainElement(interpret.Use{Typed: "height 30 cm", Meaning: "Height 300 mm", Kind: interpret.Size}))
	})

	It("lets a size word with a typo name the dimension", func() {
		// Arrange
		d := "bowl 25 cm wdie"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["diameter"]).To(BeNumerically("~", 250, tol))
	})

	It("lets a stated size win over a feature that changes the same setting", func() {
		// Arrange
		d := "tall vase 20 cm"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["height"]).To(BeNumerically("~", 200, tol))
	})

	It("uses the main dimension when the size word doesn't fit the family, and says so", func() {
		// Arrange
		d := "tray 30 cm wide"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["length"]).To(BeNumerically("~", 300, tol))
		Expect(got.Notes).To(ContainElement(`"wide" doesn't name a size of a Tray; 30 cm went to its Length.`))
	})

	It("clamps a size to the range and says so", func() {
		// Arrange
		d := "vase 2000 mm"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["height"]).To(BeNumerically("~", 400, tol))
		Expect(got.Notes).To(ContainElement("Height 2000 mm is more than the largest allowed, 400 mm; using 400 mm."))
	})

	It("leaves out a number without a unit, and says how to fix it", func() {
		// Arrange
		d := "vase 20"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["height"]).To(BeNumerically("~", 200, tol))
		Expect(got.Notes).To(ContainElement(`"20" has no unit, so it was left out. Add mm, cm or in after it.`))
	})

	It("doesn't take a size word that another size already used", func() {
		// Arrange
		d := "bowl 10 cm wide 20 cm"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["diameter"]).To(BeNumerically("~", 100, tol))
		Expect(got.Values["height"]).To(BeNumerically("~", 200, tol))
	})
})
