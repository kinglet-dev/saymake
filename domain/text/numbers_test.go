package text_test

import (
	"strings"

	"github.com/kinglet-dev/saymake/domain/text"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tokenize: numbers and limits", func() {
	It("separates a number from a unit written against it", func() {
		// Arrange
		in := "20cm tall"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"#", "cm", "tall"}))
		Expect(ts[0].Number).To(BeNumerically("~", 20, 1e-12))
	})

	It("reads decimals with a point", func() {
		// Arrange
		in := "12.75 in"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(ts[0].Number).To(BeNumerically("~", 12.75, 1e-12))
	})

	It("reads a double quote straight after a number as inches", func() {
		// Arrange
		in := `8" bowl`

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"#", "in", "bowl"}))
	})

	It("ignores a double quote anywhere else", func() {
		// Arrange
		in := `"tall" vase`

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"tall", "vase"}))
	})

	It("keeps letters beyond ASCII inside words", func() {
		// Arrange
		in := "Schälchen"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"schälchen"}))
	})

	It("reads a point after a number as the end of a sentence", func() {
		// Arrange
		in := "a vase 20. Tall"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"a", "vase", "#", "tall"}))
		Expect(ts[2].Number).To(BeNumerically("~", 20, 1e-12))
	})

	It("ends a number at a second point", func() {
		// Arrange
		in := "1.5.2"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"#", "#"}))
		Expect(ts[0].Number).To(BeNumerically("~", 1.5, 1e-12))
		Expect(ts[1].Number).To(BeNumerically("~", 2, 1e-12))
	})

	DescribeTable("refuses input that could be hostile or is unreadable",
		func(in, want string) {
			// Arrange
			// (input from the table)

			// Act
			_, err := text.Tokenize(in)

			// Assert
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("too long", strings.Repeat("a", 501), "at most 500 characters"),
		Entry("too many words", strings.Repeat("a ", 61), "at most 60 words"),
		Entry("too many digits before the point", "12345 mm", `"12345" has too many digits`),
		Entry("too many digits after the point", "1.234 mm", `"1.234" has too many digits`),
		Entry("not UTF-8", "vase \xff", "not valid text"),
	)

	It("accepts exactly the limits", func() {
		// Arrange
		in := strings.TrimSpace(strings.Repeat("ab ", 60))

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(ts).To(HaveLen(60))
	})

	It("accepts 500 characters", func() {
		// Arrange
		in := strings.Repeat("a", 500)

		// Act
		_, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("Tokenize: 2D and 3D", func() {
	It("reads 3d and 2d as words, not numbers without units", func() {
		// Arrange
		in := "3d printed vase, 2D coaster, 3D"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"3d", "printed", "vase", "2d", "coaster", "3d"}))
	})

	It("still reads a number before a word starting with d", func() {
		// Arrange
		in := "3 deep, 3dl"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"#", "deep", "#", "dl"}))
	})
})

var _ = Describe("Tokenize: breaks", func() {
	It("marks a token that follows punctuation", func() {
		// Arrange
		in := "rounded corners, 4in; tall. vase"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		var breaks []bool
		for _, t := range ts {
			breaks = append(breaks, t.AfterBreak)
		}
		Expect(breaks).To(Equal([]bool{false, false, true, false, true, true}))
	})

	It("doesn't mark a token after spaces or an apostrophe", func() {
		// Arrange
		in := "mom's  tray"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(ts[1].AfterBreak).To(BeFalse())
	})
})
