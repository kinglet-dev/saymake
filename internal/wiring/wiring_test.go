package wiring_test

import (
	"errors"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/vocab"
	"github.com/kinglet-dev/saymake/internal/wiring"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Assemble", func() {
	It("assembles the built-in parts", func() {
		// Arrange
		// (none)

		// Act
		app, err := wiring.New()

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(app.Families.IDs()).To(Equal([]string{"vessel", "tray", "flat"}))
		Expect(app.Interpreter).NotTo(BeNil())
		Expect(app.Builder).NotTo(BeNil())
	})

	DescribeTable("refuses parts that don't fit together",
		func(edit func(*wiring.Parts), want string) {
			// Arrange
			p := wiring.BuiltIn()
			edit(&p)

			// Act
			_, err := wiring.Assemble(p)

			// Assert
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
		Entry("a family twice", func(p *wiring.Parts) { p.Families = append(p.Families, p.Families[0]) }, `family "vessel" is registered twice`),
		Entry("a vocabulary that won't load", func(p *wiring.Parts) {
			p.Vocabulary = func() (vocab.Vocabulary, error) { return vocab.Vocabulary{}, errors.New("broken vocabulary") }
		}, "broken vocabulary"),
		Entry("a vocabulary for missing families", func(p *wiring.Parts) { p.Families = []family.Family{p.Families[0]} }, `family "tray" is not registered`),
		Entry("a format twice", func(p *wiring.Parts) { p.Exporters = append(p.Exporters, p.Exporters[0]) }, `format "stl" is registered twice`),
	)
})

var _ = Describe("Handler", func() {
	It("answers requests with the built-in parts", func() {
		// Arrange
		h := wiring.Handler(wiring.BuiltIn())

		// Act
		out := h.Handle([]byte(`{"op": "read", "text": "vase"}`))

		// Assert
		Expect(string(out.JSON)).To(HavePrefix(`{"ok":true`))
	})

	It("answers every request with the reason when the parts don't fit", func() {
		// Arrange
		p := wiring.BuiltIn()
		p.Exporters = append(p.Exporters, p.Exporters[0])
		h := wiring.Handler(p)

		// Act
		out := h.Handle([]byte(`{"op": "read", "text": "vase"}`))

		// Assert
		Expect(string(out.JSON)).To(ContainSubstring(`Saymake couldn't start: format \"stl\" is registered twice`))
	})
})
