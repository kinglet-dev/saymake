package interpret_test

import (
	"github.com/kinglet-dev/saymake/app/interpret"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interpreter.Read: a size word before the size that doesn't fit", func() {
	It("names the amount in the note", func() {
		// Arrange
		in, err := interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		got, err := in.Read("tray, wide of 30 cm")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Notes).To(ContainElement(`"wide" doesn't name a size of a Tray; 30 cm went to its Length.`))
	})
})

var _ = Describe("Interpreter.Read: sizes don't reach across punctuation", func() {
	It("doesn't take a size word before a comma", func() {
		// Arrange
		in, err := interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		got, err := in.Read("bowl wide, 25 cm")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["height"]).To(BeNumerically("~", 250, tol))
		Expect(got.Values["diameter"]).To(BeNumerically("~", 210, tol), `"wide" counts as the feature (+30 mm), not the size`)
	})

	It("doesn't reach across a comma among small words before the size", func() {
		// Arrange
		in, err := interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		got, err := in.Read("vase, height of, a 25 cm")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Uses).To(ContainElement(interpret.Use{Typed: "25 cm", Meaning: "Height 250 mm", Kind: interpret.Size}))
		Expect(got.Unknown).To(Equal([]string{"height"}))
	})

	It("doesn't take a size word after a comma", func() {
		// Arrange
		in, err := interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		got, err := in.Read("bowl 25 cm, wide")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["height"]).To(BeNumerically("~", 250, tol))
	})
})

var _ = Describe("Interpreter.Read: a bare size scales the shape", func() {
	var in *interpret.Interpreter

	BeforeEach(func() {
		// Arrange (shared)
		var err error
		in, err = interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())
	})

	It("scales the family's proportional settings with the main size", func() {
		// Arrange: a tray is 250 × 180 by default
		d := "tray 50 cm"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["length"]).To(BeNumerically("~", 500, tol))
		Expect(got.Values["width"]).To(BeNumerically("~", 360, tol))
	})

	It("scales by the size actually used after clamping, and clamps the others too", func() {
		// Arrange: 900 mm clamps to 500, so width scales by 2 to 360
		d := "tray 90 cm"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["width"]).To(BeNumerically("~", 360, tol))
	})

	It("lets a named size override after a bare one, without scaling again", func() {
		// Arrange
		d := "tray 50 cm, 30 cm long"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["length"]).To(BeNumerically("~", 300, tol))
		Expect(got.Values["width"]).To(BeNumerically("~", 360, tol))
	})

	It("doesn't scale anything for a named size", func() {
		// Arrange
		d := "tray 40 cm long"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["width"]).To(BeNumerically("~", 180, tol))
	})
})
