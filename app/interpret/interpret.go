// Package interpret is the use case that reads a description: it finds the
// object, the features and the sizes, and works out the starting settings,
// recording what each word did so the reading can be shown back.
package interpret

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/match"
	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/domain/text"
	"github.com/kinglet-dev/saymake/domain/vocab"
)

// Kind is what a typed word did.
type Kind string

// The kinds of use.
const (
	Object      Kind = "object"
	Feature     Kind = "feature"
	Size        Kind = "size"
	Count       Kind = "count"
	Unused      Kind = "unused"
	Unsupported Kind = "unsupported"
)

// Use is one thing the reading took from the description.
type Use struct {
	Typed, Meaning string
	Kind           Kind
}

// Interpretation is the reading of a description.
type Interpretation struct {
	Concept, Family string
	Path            []string
	Values          param.Values
	Uses            []Use
	Unknown, Notes  []string
}

type targetKind int

const (
	conceptTarget targetKind = iota
	modifierTarget
	ignoreTarget
	unsupportedTarget
)

type target struct {
	kind targetKind
	i    int
}

// Interpreter reads descriptions with one vocabulary.
type Interpreter struct {
	v           vocab.Vocabulary
	reg         *family.Registry
	words       *match.Index
	targets     []target
	counts      *match.Index // modifiers that take a count, by index
	sizes       *match.Index
	sizeWords   []string
	suggestions string
}

// New checks the vocabulary against the families and prepares the matchers.
func New(v vocab.Vocabulary, r *family.Registry) (*Interpreter, error) {
	if err := v.Validate(r); err != nil {
		return nil, err
	}
	in := &Interpreter{v: v, reg: r}
	var phrases []match.Phrase
	add := func(text string, t target, exact bool) {
		phrases = append(phrases, match.Phrase{Text: text, Ref: len(in.targets), Exact: exact})
		in.targets = append(in.targets, t)
	}
	var names []string
	for i, c := range v.Concepts {
		for _, w := range c.Words {
			add(w, target{conceptTarget, i}, false)
		}
		names = append(names, c.Words[0])
	}
	var countPhrases []match.Phrase
	for i, m := range v.Modifiers {
		for _, w := range m.Words {
			if len(m.Ops) > 0 {
				add(w, target{modifierTarget, i}, false)
			}
			if m.Count != "" {
				countPhrases = append(countPhrases, match.Phrase{Text: w, Ref: i})
			}
		}
	}
	in.counts = match.New(countPhrases)
	for _, w := range v.Ignore {
		add(w, target{ignoreTarget, 0}, true)
	}
	for i, u := range v.Unsupported {
		for _, w := range u.Words {
			add(w, target{unsupportedTarget, i}, false)
		}
	}
	in.words = match.New(phrases)
	// A typed word the vocabulary knows (anywhere) is never read as a typo
	// of a different word: "base" is a size word, not a misspelled "vase".
	var known []string
	for _, p := range phrases {
		known = append(known, strings.Fields(p.Text)...)
	}
	in.words.Protect(append(known, sortedKeys(v.Sizes)...))
	var sizePhrases []match.Phrase
	for i, w := range sortedKeys(v.Sizes) {
		sizePhrases = append(sizePhrases, match.Phrase{Text: w, Ref: i})
		in.sizeWords = append(in.sizeWords, w)
	}
	in.sizes = match.New(sizePhrases)
	in.suggestions = list(names[:min(8, len(names))])
	return in, nil
}

// Read interprets a description.
func (in *Interpreter) Read(description string) (Interpretation, error) {
	tokens, err := text.Tokenize(description)
	if err != nil {
		return Interpretation{}, err
	}
	r := &reading{in: in, tokens: tokens, used: make([]bool, len(tokens))}
	r.scanNumbers()
	r.scanWords()
	if err := r.chooseConcept(); err != nil {
		return Interpretation{}, err
	}
	r.settle()
	slices.SortStableFunc(r.uses, func(a, b placedUse) int { return a.at - b.at })
	for _, u := range r.uses {
		r.out.Uses = append(r.out.Uses, u.Use)
	}
	return r.out, nil
}

func (in *Interpreter) isIgnored(word string) bool { return slices.Contains(in.v.Ignore, word) }

func list(ws []string) string {
	if len(ws) == 1 {
		return ws[0]
	}
	return strings.Join(ws[:len(ws)-1], ", ") + " or " + ws[len(ws)-1]
}

func article(label string) string {
	if strings.ContainsRune("AEIOUaeiou", rune(label[0])) {
		return "an " + label
	}
	return "a " + label
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// label names what a target means, for notes.
func (in *Interpreter) label(t target) string {
	switch t.kind {
	case conceptTarget:
		return in.v.Concepts[t.i].Label
	case unsupportedTarget:
		return in.v.Unsupported[t.i].Words[0]
	}
	return in.v.Modifiers[t.i].Label
}

// errNoObject says what to name; if the description named only things the
// tool can't make yet, it says that first.
func errNoObject(in *Interpreter, cannot []string) error {
	if len(cannot) > 0 {
		return fmt.Errorf("%s Name an object it can make, such as %s.", strings.Join(cannot, " "), in.suggestions)
	}
	return fmt.Errorf("I couldn't tell what to make. Name an object, such as %s.", in.suggestions)
}
