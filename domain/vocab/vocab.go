// Package vocab is the vocabulary: a hierarchy of object concepts, the
// modifiers that change their settings, and size words.
package vocab

import (
	"fmt"
	"slices"

	"github.com/kinglet-dev/saymake/domain/param"
)

// OpKind is how a modifier changes a setting.
type OpKind string

// The kinds of change.
const (
	Set   OpKind = "set"
	Scale OpKind = "scale"
	Add   OpKind = "add"
)

// Op is one change to one setting.
type Op struct {
	Param string
	Kind  OpKind
	Value float64
}

// Concept is an object in the hierarchy.
type Concept struct {
	ID, Label, Parent, Family string
	Words                     []string
	Defaults                  param.Values
}

// Modifier is a word or phrase that changes settings.
type Modifier struct {
	ID, Label string
	Words     []string
	Families  []string // empty: every family
	Ops       []Op
	Count     string // "8 flutes": a number before the word sets this setting
}

// Unsupported is words for things the tool can't make yet, and what to say
// when they're used (honest instead of "not understood").
type Unsupported struct {
	Words   []string
	Message string
}

// Vocabulary is everything the interpreter knows.
type Vocabulary struct {
	Concepts    []Concept
	Modifiers   []Modifier
	Sizes       map[string]map[string]string // size word → family → setting
	Ignore      []string
	Unsupported []Unsupported
}

// Resolved is a concept with its chain worked out.
type Resolved struct {
	ID, Family string
	Path       []string // labels from the root down
	Defaults   param.Values
}

// Resolve works out a concept's family, path and merged defaults: the
// nearest family up the chain, and defaults merged from the root down, so a
// child's value replaces its parent's. The vocabulary must be valid.
func (v Vocabulary) Resolve(id string) (Resolved, error) {
	chain := v.chain(id)
	if chain == nil {
		return Resolved{}, fmt.Errorf("concept %q is not in the vocabulary", id)
	}
	r := Resolved{ID: id, Defaults: param.Values{}}
	for i := len(chain) - 1; i >= 0; i-- {
		c := chain[i]
		r.Path = append(r.Path, c.Label)
		if c.Family != "" {
			r.Family = c.Family
		}
		for k, val := range c.Defaults {
			r.Defaults[k] = val
		}
	}
	return r, nil
}

// chain returns the concept and its ancestors, nearest first, or nil if the
// id is unknown. It stops at a missing parent or a loop.
func (v Vocabulary) chain(id string) []Concept {
	var out []Concept
	for id != "" && len(out) <= len(v.Concepts) {
		c, ok := v.concept(id)
		if !ok {
			break
		}
		out = append(out, c)
		id = c.Parent
	}
	return out
}

func (v Vocabulary) concept(id string) (Concept, bool) {
	i := slices.IndexFunc(v.Concepts, func(c Concept) bool { return c.ID == id })
	if i < 0 {
		return Concept{}, false
	}
	return v.Concepts[i], true
}

// IsAncestor reports whether a is a strict ancestor of b.
func (v Vocabulary) IsAncestor(a, b string) bool {
	chain := v.chain(b)
	return slices.ContainsFunc(chain[min(1, len(chain)):], func(c Concept) bool { return c.ID == a })
}

// AppliesTo reports whether the modifier applies to a family.
func (m Modifier) AppliesTo(familyID string) bool {
	return len(m.Families) == 0 || slices.Contains(m.Families, familyID)
}
