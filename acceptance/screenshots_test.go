package acceptance_test

import (
	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/internal/wiring"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Descriptions from the page's first use that it misread, with what each
// should have made.
var _ = Describe("Descriptions from first use", func() {
	var app wiring.App

	BeforeEach(func() {
		// Arrange (shared)
		var err error
		app, err = wiring.New()
		Expect(err).NotTo(HaveOccurred())
	})

	type want struct {
		concept string
		values  map[string]float64
		atLeast map[string]float64 // settings that must be at least this
		unknown []string
		note    string
	}

	DescribeTable("each lands on what was meant",
		func(description string, w want) {
			// Arrange
			d := description

			// Act
			r, err := app.Interpreter.Read(d)
			Expect(err).NotTo(HaveOccurred())
			_, buildErr := app.Builder.Build(r.Family, r.Values, check.Context{})

			// Assert
			Expect(r.Concept).To(Equal(w.concept))
			for k, v := range w.values {
				Expect(r.Values[k]).To(BeNumerically("~", v, tol), k)
			}
			for k, v := range w.atLeast {
				Expect(r.Values[k]).To(BeNumerically(">=", v), k)
			}
			if w.unknown == nil {
				Expect(r.Unknown).To(BeEmpty())
			} else {
				Expect(r.Unknown).To(Equal(w.unknown))
			}
			if w.note != "" {
				Expect(r.Notes).To(ContainElement(ContainSubstring(w.note)))
			}
			Expect(buildErr).NotTo(HaveOccurred())
		},
		Entry(nil, "hexagonal coaster with rounded corners, 4in", want{concept: "coaster",
			values: map[string]float64{"sides": 6, "length": 101.6}, atLeast: map[string]float64{"corner": 3}}),
		Entry(nil, "a jar with chamfered brim", want{concept: "jar", values: map[string]float64{"rimStyle": 1}, atLeast: map[string]float64{"rim": 0.5}}),
		Entry(nil, "a jar with chamfered edge", want{concept: "jar", values: map[string]float64{"rimStyle": 1}, atLeast: map[string]float64{"rim": 0.5}}),
		Entry(nil, "a tray with slanted walls", want{concept: "tray", atLeast: map[string]float64{"flare": 10}}),
		Entry(nil, "a tag with the word Hello", want{concept: "tag", unknown: []string{"hello"}, note: "Text and logos aren't supported yet"}),
		Entry(nil, "bowl, 4in diameter and raised base", want{concept: "bowl", values: map[string]float64{"diameter": 101.6}, atLeast: map[string]float64{"foot": 2}}),
		Entry(nil, "cup with think handle", want{concept: "cup", atLeast: map[string]float64{"handle": 0.3}, unknown: []string{"think"},
			note: `"think" could mean "thick handle" or "thin handle"`}),
		Entry(nil, "cup with thick handle", want{concept: "cup", atLeast: map[string]float64{"handle": 0.3, "handleThick": 10}}),
	)
})
