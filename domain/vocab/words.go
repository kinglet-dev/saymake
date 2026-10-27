package vocab

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kinglet-dev/saymake/domain/family"
)

func (v Vocabulary) validateSizes(r *family.Registry) error {
	for _, w := range sortedKeys(v.Sizes) {
		if !wordRE.MatchString(w) {
			return fmt.Errorf("size word %q must be lowercase letters", w)
		}
		for _, id := range sortedKeys(v.Sizes[w]) {
			f, ok := r.Get(id)
			if !ok {
				return fmt.Errorf("size word %q: family %q is not registered", w, id)
			}
			if _, ok := f.Params().Find(v.Sizes[w][id]); !ok {
				return fmt.Errorf("size word %q: family %q has no setting %q", w, id, v.Sizes[w][id])
			}
		}
	}
	return nil
}

func checkWords(kind, id string, words []string) error {
	if len(words) == 0 {
		return fmt.Errorf("%s %q has no words", kind, id)
	}
	for _, w := range words {
		if strings.TrimSpace(w) == "" {
			return fmt.Errorf("%s %q: a word is blank", kind, id)
		}
		if !phraseRE.MatchString(w) {
			return fmt.Errorf("%s %q: %q must be lowercase letters and single spaces", kind, id, w)
		}
	}
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// checkUnique refuses a phrase listed for two different meanings: the first
// would always win, silently. A count word may also be a feature's word
// (they are matched separately).
func (v Vocabulary) checkUnique() error {
	owner := map[string]string{}
	claim := func(what string, words []string) error {
		for _, w := range words {
			if o, ok := owner[w]; ok && o != what {
				return fmt.Errorf("%q means both %s and %s", w, o, what)
			}
			owner[w] = what
		}
		return nil
	}
	for _, c := range v.Concepts {
		if err := claim(fmt.Sprintf("concept %q", c.ID), c.Words); err != nil {
			return err
		}
	}
	for _, m := range v.Modifiers {
		if len(m.Ops) == 0 {
			continue // count-only words are matched on their own
		}
		if err := claim(fmt.Sprintf("modifier %q", m.ID), m.Words); err != nil {
			return err
		}
	}
	for _, u := range v.Unsupported {
		if err := claim(fmt.Sprintf("unsupported %q", u.Message), u.Words); err != nil {
			return err
		}
	}
	return claim("an ignored word", v.Ignore)
}
