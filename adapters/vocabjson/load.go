// Package vocabjson reads the vocabulary from JSON, with limits on size and
// counts (see the threat model), and embeds the default vocabulary.
package vocabjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/domain/vocab"
)

// Limits on a vocabulary file.
const (
	MaxBytes   = 256 << 10
	MaxEntries = 400 // concepts, modifiers, size words, ignored words
	MaxWords   = 32  // words per entry, changes per modifier
)

const version = 1

type file struct {
	Version     int                          `json:"version"`
	Concepts    []concept                    `json:"concepts"`
	Modifiers   []modifier                   `json:"modifiers"`
	Sizes       map[string]map[string]string `json:"sizes"`
	Ignore      []string                     `json:"ignore"`
	Unsupported []unsupported                `json:"unsupported"`
}

type unsupported struct {
	Words   []string `json:"words"`
	Message string   `json:"message"`
}

type concept struct {
	ID       string       `json:"id"`
	Label    string       `json:"label"`
	Parent   string       `json:"parent"`
	Family   string       `json:"family"`
	Words    []string     `json:"words"`
	Defaults param.Values `json:"defaults"`
}

type modifier struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Words    []string `json:"words"`
	Families []string `json:"families"`
	Ops      []op     `json:"ops"`
	Count    string   `json:"count"`
}

type op struct {
	Param string   `json:"param"`
	Set   *float64 `json:"set"`
	Scale *float64 `json:"scale"`
	Add   *float64 `json:"add"`
}

// Load reads a vocabulary. It checks the file's form and limits; whether
// it fits the registered families is vocab.Validate's job.
func Load(r io.Reader) (vocab.Vocabulary, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxBytes+1))
	if err != nil {
		return vocab.Vocabulary{}, err
	}
	if len(data) > MaxBytes {
		return vocab.Vocabulary{}, fmt.Errorf("the vocabulary is larger than %d KB", MaxBytes>>10)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var f file
	if err := dec.Decode(&f); err != nil {
		if strings.Contains(err.Error(), "unknown field") {
			return vocab.Vocabulary{}, fmt.Errorf("the vocabulary has an %v", strings.TrimPrefix(err.Error(), "json: "))
		}
		return vocab.Vocabulary{}, fmt.Errorf("the vocabulary isn't valid JSON: %w", err)
	}
	if dec.More() {
		return vocab.Vocabulary{}, errors.New("unexpected data after the vocabulary")
	}
	if f.Version != version {
		return vocab.Vocabulary{}, fmt.Errorf("vocabulary version %d isn't supported; this build reads version %d", f.Version, version)
	}
	if err := f.checkLimits(); err != nil {
		return vocab.Vocabulary{}, err
	}
	return f.convert()
}

func (f file) checkLimits() error {
	for _, c := range []struct {
		name string
		n    int
	}{{"concepts", len(f.Concepts)}, {"modifiers", len(f.Modifiers)}, {"size words", len(f.Sizes)}, {"ignored words", len(f.Ignore)}, {"unsupported entries", len(f.Unsupported)}} {
		if c.n > MaxEntries {
			return fmt.Errorf("the vocabulary has more than %d %s", MaxEntries, c.name)
		}
	}
	for _, c := range f.Concepts {
		if len(c.Words) > MaxWords {
			return fmt.Errorf("concept %q has more than %d words", c.ID, MaxWords)
		}
	}
	for _, u := range f.Unsupported {
		if len(u.Words) > MaxWords {
			return fmt.Errorf("unsupported entry %q has more than %d words", u.Message, MaxWords)
		}
	}
	for _, m := range f.Modifiers {
		if len(m.Words) > MaxWords {
			return fmt.Errorf("modifier %q has more than %d words", m.ID, MaxWords)
		}
		if len(m.Ops) > MaxWords {
			return fmt.Errorf("modifier %q has more than %d changes", m.ID, MaxWords)
		}
	}
	return nil
}

func (f file) convert() (vocab.Vocabulary, error) {
	v := vocab.Vocabulary{Sizes: f.Sizes, Ignore: f.Ignore}
	for _, u := range f.Unsupported {
		v.Unsupported = append(v.Unsupported, vocab.Unsupported(u))
	}
	for _, c := range f.Concepts {
		v.Concepts = append(v.Concepts, vocab.Concept(c))
	}
	for _, m := range f.Modifiers {
		out := vocab.Modifier{ID: m.ID, Label: m.Label, Words: m.Words, Families: m.Families, Count: m.Count}
		for _, o := range m.Ops {
			k, val, ok := o.kind()
			if !ok {
				return vocab.Vocabulary{}, fmt.Errorf("modifier %q: each change needs exactly one of set, scale or add", m.ID)
			}
			out.Ops = append(out.Ops, vocab.Op{Param: o.Param, Kind: k, Value: val})
		}
		v.Modifiers = append(v.Modifiers, out)
	}
	return v, nil
}

func (o op) kind() (vocab.OpKind, float64, bool) {
	var k vocab.OpKind
	var val float64
	n := 0
	for _, c := range []struct {
		k vocab.OpKind
		v *float64
	}{{vocab.Set, o.Set}, {vocab.Scale, o.Scale}, {vocab.Add, o.Add}} {
		if c.v != nil {
			k, val, n = c.k, *c.v, n+1
		}
	}
	return k, val, n == 1
}
