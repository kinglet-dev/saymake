package api_test

import (
	"encoding/json"
	"testing"

	"github.com/kinglet-dev/saymake/adapters/api"
	"github.com/kinglet-dev/saymake/internal/wiring"
)

// FuzzHandle feeds arbitrary requests to the handler: it must never panic,
// always answer valid JSON, and never send a file without saying so.
func FuzzHandle(f *testing.F) {
	for _, s := range []string{
		`{"op": "read", "text": "tall fluted vase 20 cm"}`,
		`{"op": "build", "family": "tray", "values": {"length": 300, "sides": 4}}`,
		`{"op": "export", "family": "flat", "values": {}, "format": "svg", "name": "../x"}`,
		`{"op": "build", "family": "vessel", "values": {"wall": 1e308, "height": -1e308}}`,
		`{`, `null`, `{"op": 7}`,
	} {
		f.Add([]byte(s))
	}
	app, err := wiring.New()
	if err != nil {
		f.Fatal(err)
	}
	h := api.New(app.Interpreter, app.Builder, app.Families)
	f.Fuzz(func(t *testing.T, b []byte) {
		out := h.Handle(b)
		var a struct {
			OK   bool             `json:"ok"`
			File *json.RawMessage `json:"file"`
		}
		if err := json.Unmarshal(out.JSON, &a); err != nil {
			t.Fatalf("answer isn't JSON: %v: %s", err, out.JSON)
		}
		if out.File != nil && (a.File == nil || !a.OK) {
			t.Fatalf("file bytes sent without a file answer")
		}
	})
}
