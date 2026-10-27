package vocab

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"

	"github.com/kinglet-dev/saymake/domain/family"
)

// MaxDepth is the deepest a concept may sit in the hierarchy.
const MaxDepth = 12

var (
	phraseRE = regexp.MustCompile(`^\p{Ll}+( \p{Ll}+)*$`)
	wordRE   = regexp.MustCompile(`^\p{Ll}+$`)
	ignoreRE = regexp.MustCompile(`^[\p{Ll}\p{Nd}]+$`) // "3d" is a word the tokenizer makes
)

// Validate checks the vocabulary against the registered families, so that a
// typo in it is refused when it loads instead of being silently ignored.
func (v Vocabulary) Validate(r *family.Registry) error {
	if err := v.validateConcepts(r); err != nil {
		return err
	}
	if err := v.validateModifiers(r); err != nil {
		return err
	}
	if err := v.validateSizes(r); err != nil {
		return err
	}
	if err := v.checkUnique(); err != nil {
		return err
	}
	for _, w := range v.Ignore {
		if !ignoreRE.MatchString(w) {
			return fmt.Errorf("ignored word %q must be lowercase letters or digits", w)
		}
	}
	for _, u := range v.Unsupported {
		if strings.TrimSpace(u.Message) == "" {
			return errors.New("unsupported words need a message")
		}
		if err := checkWords("unsupported", u.Message, u.Words); err != nil {
			return err
		}
	}
	return nil
}

func (v Vocabulary) validateConcepts(r *family.Registry) error {
	seen := map[string]bool{}
	for _, c := range v.Concepts {
		switch {
		case c.ID == "":
			return errors.New("a concept has no id")
		case seen[c.ID]:
			return fmt.Errorf("concept %q is listed twice", c.ID)
		}
		seen[c.ID] = true
		if err := checkWords("concept", c.ID, c.Words); err != nil {
			return err
		}
	}
	for _, c := range v.Concepts {
		if err := v.validateChain(c, r); err != nil {
			return err
		}
	}
	return nil
}

func (v Vocabulary) validateChain(c Concept, r *family.Registry) error {
	if _, ok := v.concept(c.Parent); c.Parent != "" && !ok {
		return fmt.Errorf("concept %q: its parent %q is not in the vocabulary", c.ID, c.Parent)
	}
	chain := v.chain(c.ID)
	if len(chain) > len(v.Concepts) {
		return fmt.Errorf("concept %q: its parents form a loop", c.ID)
	}
	if len(chain) > MaxDepth {
		return fmt.Errorf("concept %q is more than %d levels deep", c.ID, MaxDepth)
	}
	res, _ := v.Resolve(c.ID)
	if res.Family == "" {
		return fmt.Errorf("concept %q has no family", c.ID)
	}
	f, ok := r.Get(res.Family)
	if !ok {
		return fmt.Errorf("concept %q: family %q is not registered", c.ID, res.Family)
	}
	for _, name := range sortedKeys(c.Defaults) {
		spec, ok := f.Params().Find(name)
		val := c.Defaults[name]
		switch {
		case !ok:
			return fmt.Errorf("concept %q: family %q has no setting %q", c.ID, res.Family, name)
		case math.IsNaN(val) || math.IsInf(val, 0):
			return fmt.Errorf("concept %q: default %s must be finite", c.ID, name)
		case val < spec.Min || val > spec.Max:
			return fmt.Errorf("concept %q: default %s %v is outside %v to %v", c.ID, name, val, spec.Min, spec.Max)
		}
	}
	return nil
}

func (v Vocabulary) validateModifiers(r *family.Registry) error {
	seen := map[string]bool{}
	for _, m := range v.Modifiers {
		switch {
		case m.ID == "":
			return errors.New("a modifier has no id")
		case seen[m.ID]:
			return fmt.Errorf("modifier %q is listed twice", m.ID)
		case len(m.Ops) == 0 && m.Count == "":
			return fmt.Errorf("modifier %q changes nothing", m.ID)
		}
		seen[m.ID] = true
		if err := checkWords("modifier", m.ID, m.Words); err != nil {
			return err
		}
		if err := validateModifier(m, r); err != nil {
			return fmt.Errorf("modifier %q: %w", m.ID, err)
		}
	}
	return nil
}

func validateModifier(m Modifier, r *family.Registry) error {
	for _, id := range m.Families {
		if _, ok := r.Get(id); !ok {
			return fmt.Errorf("family %q is not registered", id)
		}
	}
	names := []string{}
	for _, op := range m.Ops {
		switch {
		case op.Kind != Set && op.Kind != Scale && op.Kind != Add:
			return fmt.Errorf("%q is not set, scale or add", op.Kind)
		case math.IsNaN(op.Value) || math.IsInf(op.Value, 0):
			return errors.New("its values must be finite")
		}
		names = append(names, op.Param)
	}
	if m.Count != "" {
		names = append(names, m.Count)
	}
	for _, name := range names {
		if !anyFamilyHas(m, name, r) {
			return fmt.Errorf("no family it applies to has a setting %q", name)
		}
	}
	return nil
}

func anyFamilyHas(m Modifier, name string, r *family.Registry) bool {
	return slices.ContainsFunc(r.IDs(), func(id string) bool {
		f, _ := r.Get(id)
		_, ok := f.Params().Find(name)
		return ok && m.AppliesTo(id)
	})
}
