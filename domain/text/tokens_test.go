package text_test

import (
	"github.com/kinglet-dev/saymake/domain/text"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func words(ts []text.Token) []string {
	var out []string
	for _, t := range ts {
		if t.IsNumber {
			out = append(out, "#")
		} else {
			out = append(out, t.Text)
		}
	}
	return out
}

var _ = Describe("Tokenize", func() {
	It("lowercases words and splits on spaces and punctuation", func() {
		// Arrange
		in := "A Tall, FLUTED vase!"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"a", "tall", "fluted", "vase"}))
	})

	It("splits hyphenated words", func() {
		// Arrange
		in := "bud-vase"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"bud", "vase"}))
	})

	It("drops apostrophes inside words", func() {
		// Arrange
		in := "mom's tray"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"moms", "tray"}))
	})

	It("drops an apostrophe at the very end", func() {
		// Arrange
		in := "the dogs'"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(words(ts)).To(Equal([]string{"the", "dogs"}))
	})

	It("reads the longest numbers allowed", func() {
		// Arrange
		in := "9999.99"

		// Act
		ts, err := text.Tokenize(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(ts[0].Number).To(BeNumerically("~", 9999.99, 1e-9))
	})
})
