package interpret_test

import (
	"github.com/kinglet-dev/saymake/app/interpret"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interpreter.Read: features", func() {
	var in *interpret.Interpreter

	BeforeEach(func() {
		// Arrange (shared)
		var err error
		in, err = interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())
	})

	It("sets a value", func() {
		// Arrange
		d := "fluted vase"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["flutes"]).To(BeNumerically("~", 12, tol))
		Expect(got.Uses).To(ContainElement(interpret.Use{Typed: "fluted", Meaning: "fluted", Kind: interpret.Feature}))
	})

	It("scales a value from the object's default", func() {
		// Arrange
		d := "tall vase"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["height"]).To(BeNumerically("~", 280, tol))
	})

	It("adds to a value, and applies only the changes the family has settings for", func() {
		// Arrange
		d := "wider tray"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["width"]).To(BeNumerically("~", 210, tol))
		Expect(got.Notes).To(BeEmpty())
	})

	It("says when a feature doesn't apply to the family", func() {
		// Arrange
		d := "ribbed tray"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Notes).To(ContainElement(`"ribbed" isn't used for a Tray.`))
		Expect(got.Uses).To(ContainElement(interpret.Use{Typed: "ribbed", Meaning: "fluted", Kind: interpret.Unused}))
	})

	It("says when none of a feature's changes fit the family", func() {
		// Arrange
		d := "tall tray"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Notes).To(ContainElement(`"tall" isn't used for a Tray.`))
	})

	It("clamps a feature's result to the range and says so", func() {
		// Arrange
		d := "tall tall tall tall tall vase"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["height"]).To(BeNumerically("~", 400, tol))
		Expect(got.Notes).To(ContainElement(ContainSubstring("more than the largest allowed, 400 mm")))
	})

	It("reads a count before a word", func() {
		// Arrange
		d := "vase with 8 flutes"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["flutes"]).To(BeNumerically("~", 8, tol))
		Expect(got.Uses).To(ContainElement(interpret.Use{Typed: "8 flutes", Meaning: "Flutes 8", Kind: interpret.Count}))
	})

	It("reads a count word without a number as the feature it names", func() {
		// Arrange
		d := "vase with flutes"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["flutes"]).To(BeNumerically("~", 12, tol))
		Expect(got.Notes).To(BeEmpty())
	})

	It("says when a count doesn't apply to the family", func() {
		// Arrange
		d := "tray with 8 flutes"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Notes).To(ContainElement(`"8 flutes" isn't used for a Tray.`))
	})
})
