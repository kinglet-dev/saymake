package match_test

import (
	"github.com/kinglet-dev/saymake/domain/match"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Index.Protect", func() {
	It("lets a protected word match only exactly", func() {
		// Arrange: "base" is a word the vocabulary knows elsewhere
		idx := match.New([]match.Phrase{{Text: "vase", Ref: 1}, {Text: "raised base", Ref: 2}})
		idx.Protect([]string{"base", "raised"})

		// Act
		_, fuzzy := idx.At([]string{"base"}, 0)
		h, exact := idx.At([]string{"raised", "base"}, 0)

		// Assert
		Expect(fuzzy).To(BeFalse())
		Expect(exact).To(BeTrue())
		Expect(h.Ref).To(Equal(2))
	})

	It("still lets an unprotected word match with a typo", func() {
		// Arrange
		idx := match.New([]match.Phrase{{Text: "vase", Ref: 1}})
		idx.Protect([]string{"base"})

		// Act
		h, ok := idx.At([]string{"vaze"}, 0)

		// Assert
		Expect(ok).To(BeTrue())
		Expect(h.Ref).To(Equal(1))
	})
})

var _ = Describe("Hit.Rivals", func() {
	idx := match.New([]match.Phrase{
		{Text: "thin", Ref: 1},
		{Text: "thick", Ref: 2},
		{Text: "slim", Ref: 3},
		{Text: "skinny", Ref: 4},
	})

	It("lists the other phrases equally close to a typo", func() {
		// Arrange: "think" is one edit from both "thin" and "thick"
		ws := []string{"think"}

		// Act
		h, ok := idx.At(ws, 0)

		// Assert
		Expect(ok).To(BeTrue())
		Expect(h.Ref).To(Equal(1))
		Expect(h.Rivals).To(Equal([]int{2}))
	})

	It("lists none for an exact match", func() {
		// Arrange
		ws := []string{"thin"}

		// Act
		h, _ := idx.At(ws, 0)

		// Assert
		Expect(h.Rivals).To(BeEmpty())
	})

	It("lists none when the closest is closer than the rest", func() {
		// Arrange: "slimm" is one edit from slim, two from skinny... and far from the rest
		ws := []string{"slimm"}

		// Act
		h, _ := idx.At(ws, 0)

		// Assert
		Expect(h.Ref).To(Equal(3))
		Expect(h.Rivals).To(BeEmpty())
	})
})
