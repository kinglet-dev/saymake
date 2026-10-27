package vocabjson_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kinglet-dev/saymake/adapters/vocabjson"
	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/domain/vocab"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const small = `{
  "version": 1,
  "concepts": [
    {"id": "vessel", "label": "Vessel", "family": "vessel", "words": ["vessel"]},
    {"id": "vase", "label": "Vase", "parent": "vessel", "words": ["vase", "urn"], "defaults": {"height": 200}}
  ],
  "modifiers": [
    {"id": "tall", "label": "tall", "words": ["tall"], "ops": [{"param": "height", "scale": 1.4}]},
    {"id": "fluted", "label": "fluted", "words": ["fluted"], "families": ["vessel"], "ops": [{"param": "flutes", "set": 12}, {"param": "depth", "add": 0.02}]},
    {"id": "flute-count", "label": "flutes", "words": ["flutes"], "families": ["vessel"], "count": "flutes"}
  ],
  "sizes": {"tall": {"vessel": "height"}},
  "ignore": ["a", "with"],
  "unsupported": [{"words": ["text", "engraved name"], "message": "Text isn't supported yet."}]
}`

var _ = Describe("Load", func() {
	It("reads concepts, modifiers, size words and ignored words", func() {
		// Arrange
		in := strings.NewReader(small)

		// Act
		v, err := vocabjson.Load(in)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Concepts).To(Equal([]vocab.Concept{
			{ID: "vessel", Label: "Vessel", Family: "vessel", Words: []string{"vessel"}},
			{ID: "vase", Label: "Vase", Parent: "vessel", Words: []string{"vase", "urn"}, Defaults: param.Values{"height": 200}},
		}))
		Expect(v.Modifiers[0].Ops).To(Equal([]vocab.Op{{Param: "height", Kind: vocab.Scale, Value: 1.4}}))
		Expect(v.Modifiers[1].Ops).To(Equal([]vocab.Op{{Param: "flutes", Kind: vocab.Set, Value: 12}, {Param: "depth", Kind: vocab.Add, Value: 0.02}}))
		Expect(v.Modifiers[1].Families).To(Equal([]string{"vessel"}))
		Expect(v.Modifiers[2].Count).To(Equal("flutes"))
		Expect(v.Sizes).To(Equal(map[string]map[string]string{"tall": {"vessel": "height"}}))
		Expect(v.Ignore).To(Equal([]string{"a", "with"}))
		Expect(v.Unsupported).To(Equal([]vocab.Unsupported{{Words: []string{"text", "engraved name"}, Message: "Text isn't supported yet."}}))
	})

	DescribeTable("refuses a malformed or oversized file",
		func(in, want string) {
			// Arrange
			r := strings.NewReader(in)

			// Act
			_, err := vocabjson.Load(r)

			// Assert
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("not JSON", `{`, "the vocabulary isn't valid JSON"),
		Entry("an unknown field", `{"version": 1, "colour": "red"}`, `unknown field "colour"`),
		Entry("a wrong type", `{"version": "one"}`, "the vocabulary isn't valid JSON"),
		Entry("another version", `{"version": 2}`, "vocabulary version 2 isn't supported; this build reads version 1"),
		Entry("data after the object", `{"version": 1} {}`, "unexpected data after the vocabulary"),
		Entry("a change with no kind", `{"version": 1, "modifiers": [{"id": "m", "words": ["m"], "ops": [{"param": "h"}]}]}`,
			`modifier "m": each change needs exactly one of set, scale or add`),
		Entry("a change with two kinds", `{"version": 1, "modifiers": [{"id": "m", "words": ["m"], "ops": [{"param": "h", "set": 1, "add": 2}]}]}`,
			`modifier "m": each change needs exactly one of set, scale or add`),
		Entry("too big", `{"version": 1, "ignore": ["`+strings.Repeat("a", vocabjson.MaxBytes)+`"]}`, "the vocabulary is larger than 256 KB"),
		Entry("too many concepts", many("concepts", `{"id": "c%d", "words": ["c"]}`, vocabjson.MaxEntries+1), "more than 400 concepts"),
		Entry("too many modifiers", many("modifiers", `{"id": "m%d", "words": ["m"], "count": "x"}`, vocabjson.MaxEntries+1), "more than 400 modifiers"),
		Entry("too many ignored words", many("ignore", `"w%d"`, vocabjson.MaxEntries+1), "more than 400 ignored words"),
		Entry("too many unsupported entries", many("unsupported", `{"words": ["w"], "message": "m%d"}`, vocabjson.MaxEntries+1), "more than 400 unsupported entries"),
		Entry("too many words for an unsupported entry", `{"version": 1, "unsupported": [{"message": "m", "words": [`+words(vocabjson.MaxWords+1)+`]}]}`, `unsupported entry "m" has more than 32 words`),
		Entry("too many size words", `{"version": 1, "sizes": {`+sizes(vocabjson.MaxEntries+1)+`}}`, "more than 400 size words"),
		Entry("too many words for a concept", `{"version": 1, "concepts": [{"id": "c", "words": [`+words(vocabjson.MaxWords+1)+`]}]}`, `concept "c" has more than 32 words`),
		Entry("too many words for a modifier", `{"version": 1, "modifiers": [{"id": "m", "count": "x", "words": [`+words(vocabjson.MaxWords+1)+`]}]}`, `modifier "m" has more than 32 words`),
		Entry("too many changes for a modifier", `{"version": 1, "modifiers": [{"id": "m", "words": ["m"], "ops": [`+ops(vocabjson.MaxWords+1)+`]}]}`, `modifier "m" has more than 32 changes`),
	)

	It("passes on a read error", func() {
		// Arrange
		r := failingReader{}

		// Act
		_, err := vocabjson.Load(r)

		// Assert
		Expect(err).To(MatchError("disk error"))
	})

	It("names the same limit every time when several are exceeded", func() {
		// Arrange
		in := `{"version": 1, "concepts": [` + strings.TrimSuffix(strings.Repeat(`{"id": "c", "words": ["c"]},`, vocabjson.MaxEntries+1), ",") +
			`], "ignore": [` + strings.TrimSuffix(strings.Repeat(`"w",`, vocabjson.MaxEntries+1), ",") + `]}`

		// Act
		var msgs []string
		for range 20 {
			_, err := vocabjson.Load(strings.NewReader(in))
			msgs = append(msgs, err.Error())
		}

		// Assert
		for _, m := range msgs {
			Expect(m).To(Equal("the vocabulary has more than 400 concepts"))
		}
	})

	It("accepts a file of exactly the largest size, and entries with the most words and changes", func() {
		// Arrange
		head := `{"version": 1, "concepts": [{"id": "c", "words": [` + words(vocabjson.MaxWords) + `]}], "modifiers": [{"id": "m", "words": [` +
			words(vocabjson.MaxWords) + `], "ops": [` + ops(vocabjson.MaxWords) + `]}], "ignore": ["`
		in := head + strings.Repeat("a", vocabjson.MaxBytes-len(head)-3) + `"]}`

		// Act
		v, err := vocabjson.Load(strings.NewReader(in))

		// Assert
		Expect(len(in)).To(Equal(vocabjson.MaxBytes))
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Modifiers[0].Ops).To(HaveLen(vocabjson.MaxWords))
	})

	It("accepts the limits exactly", func() {
		// Arrange
		in := many("concepts", `{"id": "c%d", "words": ["c"]}`, vocabjson.MaxEntries)

		// Act
		v, err := vocabjson.Load(strings.NewReader(in))

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(v.Concepts).To(HaveLen(vocabjson.MaxEntries))
	})
})

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("disk error") }

func many(field, item string, n int) string {
	var parts []string
	for i := range n {
		if strings.Contains(item, "%d") {
			parts = append(parts, fmt.Sprintf(item, i))
		}
	}
	return `{"version": 1, "` + field + `": [` + strings.Join(parts, ",") + `]}`
}

func sizes(n int) string {
	var parts []string
	for i := range n {
		parts = append(parts, fmt.Sprintf(`"s%d": {}`, i))
	}
	return strings.Join(parts, ",")
}

func words(n int) string { return strings.TrimSuffix(strings.Repeat(`"w",`, n), ",") }

func ops(n int) string {
	return strings.TrimSuffix(strings.Repeat(`{"param": "h", "set": 1},`, n), ",")
}
