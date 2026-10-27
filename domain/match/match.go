// Package match finds which vocabulary phrase a run of typed words means,
// allowing a few typing mistakes.
//
// Phrases are matched greedily, longest first (the maximum-matching method
// of dictionary-based tokenization); among phrases of the same length the
// fewest edits win, then the earlier phrase. Each typed word may differ from
// its phrase word by at most text.Allowance(phrase word) edits.
package match

import (
	"strings"

	"github.com/kinglet-dev/saymake/domain/text"
)

// Phrase is a vocabulary phrase and the caller's reference for it.
type Phrase struct {
	Text  string
	Ref   int
	Exact bool // no typing mistakes allowed (e.g. ignored small words)
}

// Hit is a match.
type Hit struct {
	Ref, Len, Edits int
	Phrase          string
	Rivals          []int // refs of other phrases exactly as close (typos only)
}

type entry struct {
	words []string
	p     Phrase
}

// Index holds phrases for matching.
type Index struct {
	byLen     map[int][]entry
	maxLen    int
	protected map[string]bool
}

// New builds an index; phrases earlier in the list win ties.
func New(phrases []Phrase) *Index {
	x := &Index{byLen: map[int][]entry{}, protected: map[string]bool{}}
	for _, p := range phrases {
		ws := strings.Fields(p.Text)
		x.byLen[len(ws)] = append(x.byLen[len(ws)], entry{words: ws, p: p})
		x.maxLen = max(x.maxLen, len(ws))
	}
	return x
}

// Protect marks words the vocabulary knows elsewhere: a typed protected
// word matches only exactly, never as a typo of something else.
func (x *Index) Protect(words []string) {
	for _, w := range words {
		x.protected[w] = true
	}
}

// At matches the words starting at i.
func (x *Index) At(words []string, i int) (Hit, bool) {
	for n := min(x.maxLen, len(words)-i); n >= 1; n-- {
		if h, ok := x.best(x.byLen[n], words[i:i+n]); ok {
			return h, true
		}
	}
	return Hit{}, false
}

func (x *Index) best(entries []entry, typed []string) (Hit, bool) {
	var h Hit
	found := false
	for _, e := range entries {
		edits, ok := x.distance(e, typed)
		switch {
		case !ok:
		case !found || edits < h.Edits:
			h, found = Hit{Ref: e.p.Ref, Len: len(typed), Edits: edits, Phrase: e.p.Text}, true
		case edits == h.Edits && edits > 0:
			h.Rivals = append(h.Rivals, e.p.Ref)
		}
	}
	return h, found
}

func (x *Index) distance(e entry, typed []string) (int, bool) {
	total := 0
	for k, w := range e.words {
		allowed := text.Allowance(w)
		if e.p.Exact || x.protected[typed[k]] {
			allowed = 0
		}
		d := text.Distance(typed[k], w)
		if d > allowed {
			return 0, false
		}
		total += d
	}
	return total, true
}
