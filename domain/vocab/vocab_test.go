package vocab_test

import (
	"strings"

	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/domain/vocab"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func sample() vocab.Vocabulary {
	return vocab.Vocabulary{
		Concepts: []vocab.Concept{
			{ID: "vessel", Label: "Vessel", Family: "vessel", Words: []string{"vessel"}},
			{ID: "vase", Label: "Vase", Parent: "vessel", Words: []string{"vase"}, Defaults: param.Values{"height": 200}},
			{ID: "bud-vase", Label: "Bud vase", Parent: "vase", Words: []string{"bud vase"}, Defaults: param.Values{"height": 120}},
			{ID: "tray", Label: "Tray", Family: "tray", Words: []string{"tray"}},
		},
		Modifiers: []vocab.Modifier{
			{ID: "fluted", Label: "fluted", Words: []string{"fluted"}, Families: []string{"vessel"},
				Ops: []vocab.Op{{Param: "flutes", Kind: vocab.Set, Value: 12}}},
			{ID: "tall", Label: "tall", Words: []string{"tall"}, Ops: []vocab.Op{{Param: "height", Kind: vocab.Scale, Value: 1.4}}},
			{ID: "flute-count", Label: "flutes", Words: []string{"flutes"}, Families: []string{"vessel"}, Count: "flutes"},
		},
		Sizes:       map[string]map[string]string{"tall": {"vessel": "height"}, "long": {"tray": "length"}},
		Unsupported: []vocab.Unsupported{{Words: []string{"text", "engraved name"}, Message: "Text isn't supported yet."}},
		Ignore:      []string{"a", "with"},
	}
}

var _ = Describe("Vocabulary.Validate: ignored words", func() {
	It("accepts digits in an ignored word, like 3d", func() {
		// Arrange
		v := sample()
		v.Ignore = append(v.Ignore, "3d")

		// Act
		err := v.Validate(registry())

		// Assert
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("Vocabulary.Validate at its limits", func() {
	It("accepts a hierarchy that is one chain of every concept, exactly 12 deep", func() {
		// Arrange
		v := sample()
		v.Concepts = v.Concepts[:1]
		parent := "vessel"
		for i := range vocab.MaxDepth - 1 {
			id := "c" + strings.Repeat("x", i)
			v.Concepts = append(v.Concepts, vocab.Concept{ID: id, Label: id, Parent: parent, Words: []string{id}})
			parent = id
		}

		// Act
		err := v.Validate(registry())

		// Assert
		Expect(v.Concepts).To(HaveLen(vocab.MaxDepth))
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("Vocabulary.Validate", func() {
	It("accepts a consistent vocabulary", func() {
		// Arrange
		v := sample()

		// Act
		err := v.Validate(registry())

		// Assert
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable("refuses an inconsistent vocabulary",
		func(edit func(*vocab.Vocabulary), want string) {
			// Arrange
			v := sample()
			edit(&v)

			// Act
			err := v.Validate(registry())

			// Assert
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("concept without id", func(v *vocab.Vocabulary) { v.Concepts[0].ID = "" }, "a concept has no id"),
		Entry("repeated concept id", func(v *vocab.Vocabulary) { v.Concepts[3].ID = "vase" }, `concept "vase" is listed twice`),
		Entry("missing parent", func(v *vocab.Vocabulary) { v.Concepts[1].Parent = "pot" }, `concept "vase": its parent "pot" is not in the vocabulary`),
		Entry("no family on the chain", func(v *vocab.Vocabulary) { v.Concepts[0].Family = "" }, `concept "vessel" has no family`),
		Entry("unknown family", func(v *vocab.Vocabulary) { v.Concepts[3].Family = "lamp" }, `concept "tray": family "lamp" is not registered`),
		Entry("default for an unknown setting", func(v *vocab.Vocabulary) { v.Concepts[1].Defaults = param.Values{"spout": 1} }, `concept "vase": family "vessel" has no setting "spout"`),
		Entry("default outside the setting's range", func(v *vocab.Vocabulary) { v.Concepts[1].Defaults = param.Values{"height": 999} }, `concept "vase": default height 999 is outside 10 to 400`),
		Entry("default that isn't finite", func(v *vocab.Vocabulary) { v.Concepts[1].Defaults = param.Values{"height": inf()} }, `concept "vase": default height must be finite`),
		Entry("a cycle", func(v *vocab.Vocabulary) { v.Concepts[0].Parent = "bud-vase" }, `concept "vessel": its parents form a loop`),
		Entry("no words", func(v *vocab.Vocabulary) { v.Concepts[3].Words = nil }, `concept "tray" has no words`),
		Entry("blank word", func(v *vocab.Vocabulary) { v.Concepts[3].Words = []string{"  "} }, `concept "tray": a word is blank`),
		Entry("word not lowercase", func(v *vocab.Vocabulary) { v.Concepts[3].Words = []string{"Tray"} }, `concept "tray": "Tray" must be lowercase letters and single spaces`),
		Entry("word with a digit", func(v *vocab.Vocabulary) { v.Concepts[3].Words = []string{"tray2"} }, `concept "tray": "tray2" must be lowercase letters and single spaces`),
		Entry("word with a double space", func(v *vocab.Vocabulary) { v.Concepts[3].Words = []string{"big  tray"} }, `"big  tray" must be lowercase letters and single spaces`),
		Entry("modifier without id", func(v *vocab.Vocabulary) { v.Modifiers[0].ID = "" }, "a modifier has no id"),
		Entry("repeated modifier id", func(v *vocab.Vocabulary) { v.Modifiers[1].ID = "fluted" }, `modifier "fluted" is listed twice`),
		Entry("modifier without words", func(v *vocab.Vocabulary) { v.Modifiers[0].Words = nil }, `modifier "fluted" has no words`),
		Entry("modifier that does nothing", func(v *vocab.Vocabulary) { v.Modifiers[0].Ops = nil }, `modifier "fluted" changes nothing`),
		Entry("modifier for an unknown family", func(v *vocab.Vocabulary) { v.Modifiers[0].Families = []string{"lamp"} }, `modifier "fluted": family "lamp" is not registered`),
		Entry("op on a setting no family has", func(v *vocab.Vocabulary) { v.Modifiers[1].Ops[0].Param = "spout" }, `modifier "tall": no family it applies to has a setting "spout"`),
		Entry("op of unknown kind", func(v *vocab.Vocabulary) { v.Modifiers[1].Ops[0].Kind = "double" }, `modifier "tall": "double" is not set, scale or add`),
		Entry("op with a non-finite value", func(v *vocab.Vocabulary) { v.Modifiers[1].Ops[0].Value = inf() }, `modifier "tall": its values must be finite`),
		Entry("count for a setting no family has", func(v *vocab.Vocabulary) { v.Modifiers[2].Count = "spout" }, `modifier "flute-count": no family it applies to has a setting "spout"`),
		Entry("size word for an unknown family", func(v *vocab.Vocabulary) { v.Sizes["tall"]["lamp"] = "height" }, `size word "tall": family "lamp" is not registered`),
		Entry("size word for an unknown setting", func(v *vocab.Vocabulary) { v.Sizes["long"]["tray"] = "girth" }, `size word "long": family "tray" has no setting "girth"`),
		Entry("size word not lowercase", func(v *vocab.Vocabulary) { v.Sizes["Wide"] = map[string]string{"tray": "length"} }, `size word "Wide" must be lowercase letters`),
		Entry("ignored word not lowercase", func(v *vocab.Vocabulary) { v.Ignore = []string{"The"} }, `ignored word "The" must be lowercase letters or digits`),
		Entry("a hierarchy too deep", deepChain, "more than 12 levels deep"),
		Entry("an ignored word that is also a concept's", func(v *vocab.Vocabulary) { v.Ignore = append(v.Ignore, "tray") }, `"tray" means both concept "tray" and an ignored word`),
		Entry("a phrase for two concepts", func(v *vocab.Vocabulary) { v.Concepts[3].Words = []string{"vase"} }, `"vase" means both concept "vase" and concept "tray"`),
		Entry("a phrase for a concept and a modifier", func(v *vocab.Vocabulary) { v.Modifiers[1].Words = []string{"tray"} }, `"tray" means both concept "tray" and modifier "tall"`),
		Entry("a phrase for a modifier and an unsupported entry", func(v *vocab.Vocabulary) { v.Unsupported[0].Words = []string{"fluted"} },
			`"fluted" means both modifier "fluted" and unsupported "Text isn't supported yet."`),
		Entry("unsupported words without a message", func(v *vocab.Vocabulary) { v.Unsupported[0].Message = " " }, "unsupported words need a message"),
		Entry("unsupported entry without words", func(v *vocab.Vocabulary) { v.Unsupported[0].Words = nil }, `unsupported "Text isn't supported yet." has no words`),
		Entry("unsupported word not lowercase", func(v *vocab.Vocabulary) { v.Unsupported[0].Words = []string{"Text"} }, `"Text" must be lowercase letters and single spaces`),
	)
})

func inf() float64 {
	var zero float64
	return 1 / zero
}

func deepChain(v *vocab.Vocabulary) {
	parent := "vase"
	for i := range 12 {
		id := "deep" + strings.Repeat("x", i)
		v.Concepts = append(v.Concepts, vocab.Concept{ID: id, Label: id, Parent: parent, Words: []string{"deep"}})
		parent = id
	}
}
