package interpret

import (
	"fmt"
	"slices"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/domain/vocab"
)

// chooseConcept picks the object: concepts that are ancestors of another
// one named are dropped (the specific one wins), then the last one named.
func (r *reading) chooseConcept() error {
	var named []wordHit
	for _, h := range r.hits {
		if h.t.kind == conceptTarget {
			named = append(named, h)
		}
	}
	v := r.in.v
	kept := slices.DeleteFunc(slices.Clone(named), func(a wordHit) bool {
		return slices.ContainsFunc(named, func(b wordHit) bool {
			return v.IsAncestor(v.Concepts[a.t.i].ID, v.Concepts[b.t.i].ID)
		})
	})
	if len(kept) == 0 {
		return errNoObject(r.in, r.cannot)
	}
	chosen := kept[len(kept)-1]
	c := v.Concepts[chosen.t.i]
	res, _ := v.Resolve(c.ID) // the vocabulary was validated in New
	r.out.Concept, r.out.Family, r.out.Path = c.ID, res.Family, res.Path
	r.addUse(chosen.typed, c.Label, Object, chosen.at)
	for _, h := range kept[:len(kept)-1] {
		r.out.Notes = append(r.out.Notes, fmt.Sprintf("%q also names an object; making %s, the last one named.", h.typed, article(c.Label)))
	}
	return nil
}

// settle works out the values: family defaults, then the concept chain's,
// then features in the order typed, then counts, then stated sizes.
func (r *reading) settle() {
	f, _ := r.in.reg.Get(r.out.Family)
	res, _ := r.in.v.Resolve(r.out.Concept)
	ps := f.Params()
	r.out.Values, _, _ = ps.Apply(ps.Defaults(), res.Defaults) // checked by vocab.Validate
	label := r.objectLabel()
	for _, h := range r.hits {
		if h.t.kind == modifierTarget {
			r.applyModifier(ps, r.in.v.Modifiers[h.t.i], h, label)
		}
	}
	for _, c := range r.counts {
		m := r.in.v.Modifiers[c.modifier]
		spec, ok := ps.Find(m.Count)
		if !ok || !m.AppliesTo(f.ID()) {
			r.unused(c.typed, m.Label, label, c.at)
			continue
		}
		got := r.set(ps, m.Count, c.n)
		r.addUse(c.typed, spec.Label+" "+spec.Format(got), Count, c.at)
	}
	for _, q := range r.quantities {
		r.applyQuantity(f, q, label)
	}
	r.out.Unknown = r.unknown
}

func (r *reading) applyModifier(ps param.Set, m vocab.Modifier, h wordHit, label string) {
	applied := false
	for _, op := range m.Ops {
		if _, ok := ps.Find(op.Param); !ok || !m.AppliesTo(r.out.Family) {
			continue
		}
		cur := r.out.Values[op.Param]
		next := map[vocab.OpKind]float64{vocab.Set: op.Value, vocab.Scale: cur * op.Value, vocab.Add: cur + op.Value}[op.Kind]
		r.set(ps, op.Param, next)
		applied = true
	}
	if !applied {
		r.unused(h.typed, m.Label, label, h.at)
		return
	}
	r.addUse(h.typed, m.Label, Feature, h.at)
}

func (r *reading) applyQuantity(f family.Family, q quantity, label string) {
	name := f.MainSize()
	if q.sizeWord != "" {
		if p, ok := r.in.v.Sizes[q.sizeWord][f.ID()]; ok {
			name = p
		} else {
			spec, _ := f.Params().Find(name)
			r.out.Notes = append(r.out.Notes, fmt.Sprintf("%q doesn't name a size of %s; %s went to its %s.",
				q.sizeWord, article(label), q.amount, spec.Label))
		}
	}
	spec, _ := f.Params().Find(name)
	before := r.out.Values[name]
	got := r.set(f.Params(), name, q.mm)
	r.addUse(q.typed, spec.Label+" "+spec.Format(got), Size, q.at)
	if q.sizeWord == "" {
		for _, p := range f.Proportional() {
			r.set(f.Params(), p, r.out.Values[p]*got/before)
		}
	}
}

// set applies one value, clamped, records any clamp note, and returns the
// value used.
func (r *reading) set(ps param.Set, name string, v float64) float64 {
	vals, notes, _ := ps.Apply(r.out.Values, param.Values{name: v}) // name is known and v finite
	r.out.Values = vals
	r.out.Notes = append(r.out.Notes, notes...)
	return vals[name]
}

func (r *reading) unused(typed, meaning, label string, at int) {
	r.out.Notes = append(r.out.Notes, fmt.Sprintf("%q isn't used for %s.", typed, article(label)))
	r.addUse(typed, meaning, Unused, at)
}

func (r *reading) addUse(typed, meaning string, k Kind, at int) {
	r.uses = append(r.uses, placedUse{Use{Typed: typed, Meaning: meaning, Kind: k}, at})
}

func (r *reading) objectLabel() string {
	i := slices.IndexFunc(r.in.v.Concepts, func(c vocab.Concept) bool { return c.ID == r.out.Concept })
	return r.in.v.Concepts[i].Label
}
