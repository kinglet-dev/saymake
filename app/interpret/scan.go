package interpret

import (
	"fmt"
	"github.com/kinglet-dev/saymake/domain/text"
	"math"
	"strings"
)

type placedUse struct {
	Use
	at int
}

type quantity struct {
	mm       float64
	sizeWord string // "" for the main size
	amount   string // the number and unit as typed, e.g. "30 cm"
	typed    string
	at       int
}

type count struct {
	modifier int
	n        float64
	typed    string
	at       int
}

type wordHit struct {
	t     target
	typed string
	at    int
}

type reading struct {
	in         *Interpreter
	tokens     []text.Token
	used       []bool
	quantities []quantity
	counts     []count
	hits       []wordHit
	unknown    []string
	cannot     []string // messages for things it can't make yet
	uses       []placedUse
	out        Interpretation
}

// scanNumbers takes sizes ("20 cm tall", "height of 20 cm") and counts
// ("8 flutes") out of the tokens. A number with neither is left out.
func (r *reading) scanNumbers() {
	for i, t := range r.tokens {
		if !t.IsNumber {
			continue
		}
		r.used[i] = true
		if i+1 < len(r.tokens) {
			if unit, ok := text.Millimeters(r.tokens[i+1].Text); ok {
				r.used[i+1] = true
				r.addQuantity(i, i+1, math.Round(t.Number*unit*100)/100)
				continue
			}
			if m, ok := r.countAt(i + 1); ok {
				r.used[i+1] = true
				r.counts = append(r.counts, count{modifier: m, n: t.Number, typed: r.span(i, i+1), at: i})
				continue
			}
		}
		r.out.Notes = append(r.out.Notes, fmt.Sprintf("%q has no unit, so it was left out. Add mm, cm or in after it.", t.Text))
	}
}

// countAt matches a count word ("flutes"). A typo may still match, unless
// the word is exactly some other word in the vocabulary ("fluted").
func (r *reading) countAt(i int) (int, bool) {
	ws := []string{r.tokens[i].Text}
	h, ok := r.in.counts.At(ws, 0)
	if ok && h.Edits > 0 && r.exactWord(ws) {
		return 0, false
	}
	return h.Ref, ok
}

func (r *reading) addQuantity(num, unit int, mm float64) {
	q := quantity{mm: mm, amount: r.span(num, unit), typed: r.span(num, unit), at: num}
	if j := unit + 1; j < len(r.tokens) && !r.used[j] && !r.tokens[j].AfterBreak {
		if w, ok := r.sizeWordAt(j); ok {
			r.used[j] = true
			q.sizeWord, q.typed = w, r.span(num, j)
			r.quantities = append(r.quantities, q)
			return
		}
	}
	j := num - 1
	for j >= 0 && !r.used[j] && !r.tokens[j+1].AfterBreak && r.in.isIgnored(r.tokens[j].Text) {
		j--
	}
	if j >= 0 && !r.used[j] && !r.tokens[j+1].AfterBreak {
		if w, ok := r.sizeWordAt(j); ok {
			for k := j; k < num; k++ {
				r.used[k] = true
			}
			q.sizeWord, q.typed, q.at = w, r.span(j, unit), j
		}
	}
	r.quantities = append(r.quantities, q)
}

// sizeWordAt matches a size word ("tall", "wide"). A typo may still match,
// unless the word is exactly some other word in the vocabulary ("vase" is
// not a typo for "base").
func (r *reading) sizeWordAt(i int) (string, bool) {
	ws := []string{r.tokens[i].Text}
	h, ok := r.in.sizes.At(ws, 0)
	if !ok || h.Edits > 0 && r.exactWord(ws) {
		return "", false
	}
	return r.in.sizeWords[h.Ref], true
}

func (r *reading) exactWord(ws []string) bool {
	w, ok := r.in.words.At(ws, 0)
	return ok && w.Edits == 0
}

func (r *reading) span(from, to int) string {
	var ws []string
	for _, t := range r.tokens[from : to+1] {
		ws = append(ws, t.Text)
	}
	return strings.Join(ws, " ")
}

// scanWords matches the words left after numbers to concepts, modifiers and
// ignored words, a run of unused words at a time.
func (r *reading) scanWords() {
	for i := 0; i < len(r.tokens); {
		if r.used[i] {
			i++
			continue
		}
		end := i
		for end < len(r.tokens) && !r.used[end] {
			end++
		}
		r.scanRun(i, end)
		i = end
	}
}

func (r *reading) scanRun(from, to int) {
	var ws []string
	for _, t := range r.tokens[from:to] {
		ws = append(ws, t.Text)
	}
	for k := 0; k < len(ws); {
		h, ok := r.in.words.At(ws, k)
		if !ok {
			r.unknown = append(r.unknown, ws[k])
			k++
			continue
		}
		typed := strings.Join(ws[k:k+h.Len], " ")
		switch t := r.in.targets[h.Ref]; {
		case r.ambiguous(h, ws[k:k+h.Len], from+k):
		case t.kind == unsupportedTarget:
			r.unsupported(t.i, typed, from+k)
		case t.kind != ignoreTarget:
			r.hits = append(r.hits, wordHit{t: t, typed: typed, at: from + k})
		}
		k += h.Len
	}
}
