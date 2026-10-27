package match_test

import (
	"github.com/kinglet-dev/saymake/domain/match"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Index.At", func() {
	idx := match.New([]match.Phrase{
		{Text: "vase", Ref: 1},
		{Text: "bud vase", Ref: 2},
		{Text: "narrow neck", Ref: 3},
		{Text: "narrow", Ref: 4},
		{Text: "the", Ref: 5, Exact: true},
		{Text: "with", Ref: 6, Exact: true},
		{Text: "vast", Ref: 7},
		{Text: "cup", Ref: 8},
	})

	It("matches a word exactly", func() {
		// Arrange
		ws := []string{"a", "vase"}

		// Act
		h, ok := idx.At(ws, 1)

		// Assert
		Expect(ok).To(BeTrue())
		Expect(h).To(Equal(match.Hit{Ref: 1, Len: 1, Edits: 0, Phrase: "vase"}))
	})

	It("prefers the longest phrase that matches", func() {
		// Arrange
		ws := []string{"bud", "vase"}

		// Act
		h, ok := idx.At(ws, 0)

		// Assert
		Expect(ok).To(BeTrue())
		Expect(h.Ref).To(Equal(2))
		Expect(h.Len).To(Equal(2))
	})

	It("matches a phrase with a typo in one word", func() {
		// Arrange
		ws := []string{"narow", "neck"}

		// Act
		h, ok := idx.At(ws, 0)

		// Assert
		Expect(ok).To(BeTrue())
		Expect(h).To(Equal(match.Hit{Ref: 3, Len: 2, Edits: 1, Phrase: "narrow neck"}))
	})

	It("falls back to a shorter phrase when the longer one doesn't match", func() {
		// Arrange
		ws := []string{"narrow", "base"}

		// Act
		h, ok := idx.At(ws, 0)

		// Assert
		Expect(ok).To(BeTrue())
		Expect(h.Ref).To(Equal(4))
	})

	It("picks the closest of several fuzzy matches", func() {
		// Arrange
		ws := []string{"vase"}

		// Act
		h, ok := idx.At(ws, 0)

		// Assert
		Expect(ok).To(BeTrue())
		Expect(h.Ref).To(Equal(1), "vase is 0 edits away, vast is 1")
	})

	It("breaks a tie by the order phrases were given", func() {
		// Arrange
		ws := []string{"vasx"}

		// Act
		h, ok := idx.At(ws, 0)

		// Assert
		Expect(ok).To(BeTrue())
		Expect(h.Ref).To(Equal(1), "vase and vast are both 1 edit away; vase came first")
	})

	It("matches an exact-only phrase only exactly", func() {
		// Arrange
		ws := []string{"wiht", "with"}

		// Act
		_, okTypo := idx.At(ws, 0)
		h, okExact := idx.At(ws, 1)

		// Assert
		Expect(okTypo).To(BeFalse())
		Expect(okExact).To(BeTrue())
		Expect(h.Ref).To(Equal(6))
	})

	It("allows no typo in a short word", func() {
		// Arrange
		ws := []string{"cap"}

		// Act
		_, ok := idx.At(ws, 0)

		// Assert
		Expect(ok).To(BeFalse())
	})

	It("finds nothing for an unrelated word", func() {
		// Arrange
		ws := []string{"lamp"}

		// Act
		_, ok := idx.At(ws, 0)

		// Assert
		Expect(ok).To(BeFalse())
	})

	It("doesn't read a phrase past the end of the words", func() {
		// Arrange
		ws := []string{"bud"}

		// Act
		_, ok := idx.At(ws, 0)

		// Assert
		Expect(ok).To(BeFalse())
	})
})
