package interpret_test

import (
	"github.com/kinglet-dev/saymake/app/interpret"
	"github.com/kinglet-dev/saymake/domain/vocab"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interpreter.Read: saying what it can't use", func() {
	var in *interpret.Interpreter

	BeforeEach(func() {
		// Arrange (shared)
		var err error
		in, err = interpret.New(vocabulary(), registry())
		Expect(err).NotTo(HaveOccurred())
	})

	It("doesn't take a word it knows for a typo of another", func() {
		// Arrange: "base" is a size word, one edit from "vase"
		d := "bowl with a base"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Concept).To(Equal("bowl"))
		Expect(got.Unknown).To(Equal([]string{"base"}))
	})

	It("asks instead of guessing when a typo is as close to two meanings", func() {
		// Arrange: "think" is one edit from "thin" and from "thick"
		d := "think vase"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Values["diameter"]).To(BeNumerically("~", 100, tol))
		Expect(got.Unknown).To(Equal([]string{"think"}))
		Expect(got.Notes).To(ContainElement(`"think" could mean "narrow" or "thick walls"; type the one you meant.`))
	})

	It("names a word it can't make yet among the meanings a typo could have", func() {
		// Arrange: "glase" is one edit from the feature "glass" and the
		// unsupported "glaze"
		v := vocabulary()
		v.Modifiers = append(v.Modifiers, vocab.Modifier{ID: "glass", Label: "glass", Words: []string{"glass"},
			Ops: []vocab.Op{{Param: "diameter", Kind: vocab.Set, Value: 50}}})
		v.Unsupported = append(v.Unsupported, vocab.Unsupported{Words: []string{"glaze"}, Message: "Glazes aren't supported."})
		in, err := interpret.New(v, registry())
		Expect(err).NotTo(HaveOccurred())

		// Act
		got, err := in.Read("glase vase")

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Notes).To(ContainElement(`"glase" could mean "glass" or "glaze"; type the one you meant.`))
	})

	It("keeps the sure words of a two-word phrase when one word is a tie", func() {
		// Arrange: "think rim" is as close to "thick rim" as to "thin rim"
		d := "vase, think rim"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Unknown).To(Equal([]string{"think"}))
		Expect(got.Notes).To(ContainElement(`"think" could mean "thick rim" or "thin rim"; type the one you meant.`))
		Expect(got.Values["flutes"]).To(BeNumerically("~", 2, tol), `"rim" still counts`)
		Expect(got.Values["diameter"]).To(BeNumerically("~", 100, tol))
	})

	It("leads with what it can't make when that's all the description names", func() {
		// Arrange
		d := "an engraved name"

		// Act
		_, err := in.Read(d)

		// Assert
		Expect(err).To(MatchError("Text isn't supported yet. Name an object it can make, such as vessel, vase, bud vase, bowl or tray."))
	})

	It("reads again every sure word before a tie in a longer phrase", func() {
		// Arrange: "big wide think" ties between two three-word phrases
		d := "vase big wide think"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Unknown).To(Equal([]string{"think"}))
		Expect(got.Values["height"]).To(BeNumerically("~", 300, tol), `"big" still counts`)
		Expect(got.Values["diameter"]).To(BeNumerically("~", 130, tol), `"wide" still counts`)
	})

	It("says when a word asks for something it can't make yet", func() {
		// Arrange
		d := "vase with the word hi and engraved name"

		// Act
		got, err := in.Read(d)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Concept).To(Equal("vase"))
		Expect(got.Unknown).To(Equal([]string{"hi"}))
		Expect(got.Notes).To(Equal([]string{"Text isn't supported yet."}), "said once")
		Expect(got.Uses).To(ContainElement(interpret.Use{Typed: "word", Meaning: "Text isn't supported yet.", Kind: interpret.Unsupported}))
		Expect(got.Uses).To(ContainElement(interpret.Use{Typed: "engraved name", Meaning: "Text isn't supported yet.", Kind: interpret.Unsupported}))
	})
})
