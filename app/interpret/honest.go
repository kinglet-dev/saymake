package interpret

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/kinglet-dev/saymake/domain/match"
)

// ambiguous reports a typo that is as close to two different meanings, and
// leaves the word unused rather than guess. In a phrase, only the typed
// words that differ from the phrase are in doubt; the rest are read again
// on their own ("think handle": "handle" still counts).
func (r *reading) ambiguous(h match.Hit, ws []string, at int) bool {
	t := r.in.targets[h.Ref]
	labels := []string{r.in.label(t)}
	for _, ref := range h.Rivals {
		if rt := r.in.targets[ref]; rt != t && !slices.Contains(labels, r.in.label(rt)) {
			labels = append(labels, r.in.label(rt))
		}
	}
	if len(labels) == 1 {
		return false
	}
	slices.Sort(labels)
	for i, l := range labels {
		labels[i] = strconv.Quote(l)
	}
	phrase := strings.Fields(h.Phrase)
	var doubt []string
	sure := -1                      // start of the current run of words typed as in the phrase
	for i := 0; i <= len(ws); i++ { // one past the end closes the last run
		if i < len(ws) && ws[i] == phrase[i] {
			if sure < 0 {
				sure = i
			}
			continue
		}
		if sure >= 0 {
			r.scanRun(at+sure, at+i)
			sure = -1
		}
		if i < len(ws) {
			doubt = append(doubt, ws[i])
		}
	}
	r.unknown = append(r.unknown, doubt...)
	r.out.Notes = append(r.out.Notes, fmt.Sprintf("%q could mean %s; type the one you meant.", strings.Join(doubt, " "), list(labels)))
	return true
}

// unsupported records a word for something the tool can't make yet, saying
// so once per kind of thing.
func (r *reading) unsupported(i int, typed string, at int) {
	msg := r.in.v.Unsupported[i].Message
	if !slices.Contains(r.cannot, msg) {
		r.cannot = append(r.cannot, msg)
		r.out.Notes = append(r.out.Notes, msg)
	}
	r.addUse(typed, msg, Unsupported, at)
}
