package text_test

import (
	"github.com/kinglet-dev/saymake/domain/text"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Distance (optimal string alignment)", func() {
	DescribeTable("counts the fewest edits",
		func(a, b string, want int) {
			// Arrange
			// (words from the table)

			// Act
			d := text.Distance(a, b)

			// Assert
			Expect(d).To(Equal(want))
		},
		Entry("same word", "vase", "vase", 0),
		Entry("one substitution", "vaze", "vase", 1),
		Entry("one insertion", "vasse", "vase", 1),
		Entry("one deletion", "vse", "vase", 1),
		Entry("one transposition", "vsae", "vase", 1),
		Entry("empty and a word", "", "tray", 4),
		Entry("a word and empty", "tray", "", 4),
		Entry("unrelated words", "bowl", "tray", 4),
		Entry("letters beyond ASCII count once", "plätte", "platte", 1),
		Entry("OSA does not edit a transposed pair again", "ca", "abc", 3),
	)
})

var _ = Describe("Allowance", func() {
	DescribeTable("allows more edits for longer words",
		func(word string, want int) {
			// Arrange
			// (word from the table)

			// Act
			n := text.Allowance(word)

			// Assert
			Expect(n).To(Equal(want))
		},
		Entry("3 letters: exact only", "cup", 0),
		Entry("4 letters: one edit", "vase", 1),
		Entry("7 letters: one edit", "planter", 1),
		Entry("8 letters: two edits", "fluteing", 2),
		Entry("letters, not bytes", "äöü", 0),
	)
})
