package vocabjson_test

import (
	"bytes"
	"testing"

	"github.com/kinglet-dev/saymake/adapters/vocabjson"
)

// FuzzLoad feeds arbitrary bytes to the loader: it must never panic, and a
// vocabulary it accepts must respect the limits.
func FuzzLoad(f *testing.F) {
	f.Add([]byte(small))
	f.Add([]byte(`{"version": 1}`))
	f.Add([]byte(`{"version": 1, "modifiers": [{"id": "m", "ops": [{"param": "h", "set": 1, "add": 2}]}]}`))
	f.Add([]byte(`[]`))
	f.Fuzz(func(t *testing.T, b []byte) {
		v, err := vocabjson.Load(bytes.NewReader(b))
		if err != nil {
			return
		}
		if len(v.Concepts) > vocabjson.MaxEntries || len(v.Modifiers) > vocabjson.MaxEntries {
			t.Fatalf("accepted %d concepts and %d modifiers", len(v.Concepts), len(v.Modifiers))
		}
	})
}
