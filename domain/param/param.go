// Package param describes a shape family's settings: their ranges, defaults
// and units, and how values are checked against them.
package param

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
)

// Spec describes one setting.
type Spec struct {
	Name, Label, Unit       string
	Min, Max, Default, Step float64
	Whole                   bool // whole numbers only (e.g. a count)
}

// Clamp limits v to the range (rounding whole-number settings) and reports
// whether the range changed it.
func (s Spec) Clamp(v float64) (float64, bool) {
	if s.Whole {
		v = math.Round(v)
	}
	switch {
	case v < s.Min:
		return s.Min, true
	case v > s.Max:
		return s.Max, true
	}
	return v, false
}

// Format writes a value with the spec's unit, e.g. "200 mm".
func (s Spec) Format(v float64) string {
	n := strconv.FormatFloat(v, 'f', -1, 64)
	if s.Unit == "" {
		return n
	}
	return n + " " + s.Unit
}

// Values maps setting names to values.
type Values map[string]float64

// Set is a family's settings, in display order.
type Set []Spec

// Find returns the spec with the given name.
func (s Set) Find(name string) (Spec, bool) {
	for _, p := range s {
		if p.Name == name {
			return p, true
		}
	}
	return Spec{}, false
}

// Defaults returns every setting's default value.
func (s Set) Defaults() Values {
	d := make(Values, len(s))
	for _, p := range s {
		d[p.Name] = p.Default
	}
	return d
}

// Validate checks that the set itself is well formed.
func (s Set) Validate() error {
	seen := map[string]bool{}
	for _, p := range s {
		if err := p.validate(); err != nil {
			return err
		}
		if seen[p.Name] {
			return fmt.Errorf("setting %q is listed twice", p.Name)
		}
		seen[p.Name] = true
	}
	return nil
}

func (p Spec) validate() error {
	switch {
	case p.Name == "":
		return errors.New("a setting has no name")
	case !finite(p.Min) || !finite(p.Max) || !finite(p.Default) || !finite(p.Step):
		return fmt.Errorf("setting %q: its numbers must be finite", p.Name)
	case p.Min > p.Max:
		return fmt.Errorf("setting %q: the range %v to %v is empty", p.Name, p.Min, p.Max)
	case p.Default < p.Min || p.Default > p.Max:
		return fmt.Errorf("setting %q: default %v is outside %v to %v", p.Name, p.Default, p.Min, p.Max)
	case p.Step <= 0:
		return fmt.Errorf("setting %q: step must be positive", p.Name)
	}
	return nil
}

// Apply returns base with changes applied, each checked and clamped, plus a
// note for every value the range changed. base is not modified.
func (s Set) Apply(base, changes Values) (Values, []string, error) {
	names := make([]string, 0, len(changes))
	for name := range changes {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		if _, ok := s.Find(name); !ok {
			return nil, nil, fmt.Errorf("unknown setting %q", name)
		}
	}
	out := make(Values, len(base))
	for k, v := range base {
		out[k] = v
	}
	var notes []string
	for _, p := range s {
		v, ok := changes[p.Name]
		if !ok {
			continue
		}
		if !finite(v) {
			return nil, nil, fmt.Errorf("%s must be a finite number", p.Label)
		}
		got, clamped := p.Clamp(v)
		if clamped {
			notes = append(notes, clampNote(p, v, got))
		}
		out[p.Name] = got
	}
	return out, notes, nil
}

func clampNote(p Spec, asked, got float64) string {
	limit := "more than the largest"
	if got == p.Min {
		limit = "less than the smallest"
	}
	return fmt.Sprintf("%s %s is %s allowed, %s; using %s.", p.Label, p.Format(asked), limit, p.Format(got), p.Format(got))
}

func finite(f float64) bool { return !math.IsNaN(f) && !math.IsInf(f, 0) }

// Require returns an error naming the first setting v has no value for.
func (s Set) Require(v Values) error {
	for _, p := range s {
		if _, ok := v[p.Name]; !ok {
			return fmt.Errorf("setting %q is missing", p.Name)
		}
	}
	return nil
}
