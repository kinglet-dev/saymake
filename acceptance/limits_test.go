package acceptance_test

import (
	"math/rand/v2"
	"unicode"

	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/domain/param"
	"github.com/kinglet-dev/saymake/internal/wiring"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// At and around every limit, each family must either refuse with a message
// a person can act on, or build a shape that passes the solid check and
// stays within the triangle budget. Settings are drawn from each range's
// minimum, maximum, default and points just inside them (a structured
// search rather than purely random values), with a fixed seed so runs repeat.
var _ = Describe("Every family at the limits of its settings", func() {
	It("refuses cleanly or builds a sound shape", func() {
		// Arrange
		app, err := wiring.New()
		Expect(err).NotTo(HaveOccurred())
		rng := rand.New(rand.NewPCG(2026, 10))
		pick := func(p param.Spec) float64 {
			span := p.Max - p.Min
			return []float64{p.Min, p.Max, p.Default, p.Min + span*0.01, p.Max - span*0.01, p.Min + span*rng.Float64()}[rng.IntN(6)]
		}
		built, refused := 0, 0

		// Act and assert, for each case
		for _, id := range app.Families.IDs() {
			f, _ := app.Families.Get(id)
			cases := []param.Values{}
			for _, p := range f.Params() {
				for _, v := range []float64{p.Min, p.Max} {
					c := f.Params().Defaults()
					c[p.Name] = v
					cases = append(cases, c)
				}
			}
			for range 400 {
				c := param.Values{}
				for _, p := range f.Params() {
					c[p.Name] = pick(p)
				}
				cases = append(cases, c)
			}
			for _, c := range cases {
				r, err := app.Builder.Build(id, c, check.Context{})
				if err != nil {
					refused++
					msg := []rune(err.Error())
					Expect(unicode.IsUpper(msg[0])).To(BeTrue(), "%s %v: %v", id, c, err)
					Expect(err.Error()).To(MatchRegexp(`\. .+\.$`), "a refusal says what to do: %s %v: %v", id, c, err)
					continue
				}
				built++
				Expect(check.Blocking(r.Findings)).To(BeFalse(), "%s %v: %v", id, c, r.Findings)
				Expect(r.Formats).NotTo(BeEmpty(), "%s %v", id, c)
				if r.Shape.Mesh != nil {
					Expect(len(r.Shape.Mesh.Triangles)).To(BeNumerically("<=", 270_000))
				}
			}
		}
		GinkgoWriter.Printf("built %d, refused %d\n", built, refused)
		// Each setting at an extreme is another way to be impossible, so with
		// more settings more random combinations refuse; a third building is
		// the floor that would show the families refusing too eagerly.
		Expect(built).To(BeNumerically(">", (built+refused)/3), "a good share of combinations should build")
	})
})
