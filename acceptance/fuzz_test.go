package acceptance_test

import (
	"testing"

	"github.com/kinglet-dev/saymake/internal/wiring"
)

// FuzzRead feeds arbitrary descriptions to the real interpreter: it must
// never panic, and any reading it gives must keep every setting in range.
func FuzzRead(f *testing.F) {
	for _, s := range []string{
		"a tall fluted vase, 20 cm", `spiral vase 8" tall`, "tray 50 cm, 30 cm long", "bowl 10 cm wide 20 cm",
		"vase with 8 flutes", "2 fluted vase", "1.5.2 cm", "9999.99 mm", "", "ñandú vase", "vase 20.", "\xff",
		"height of 30cm vase", "tag with hanging hole 5 mm", "6 sided 6 sided 6 sided bowl",
	} {
		f.Add(s)
	}
	app, err := wiring.New()
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, s string) {
		r, err := app.Interpreter.Read(s)
		if err != nil {
			return
		}
		fam, ok := app.Families.Get(r.Family)
		if !ok {
			t.Fatalf("reading of %q names unknown family %q", s, r.Family)
		}
		for _, p := range fam.Params() {
			v, ok := r.Values[p.Name]
			if !ok || v < p.Min || v > p.Max {
				t.Fatalf("reading of %q gives %s = %v, outside %v to %v", s, p.Name, v, p.Min, p.Max)
			}
		}
	})
}
