package build_test

import (
	"bytes"
	"errors"

	"github.com/kinglet-dev/saymake/app/build"
	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/domain/param"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Builder.Build", func() {
	var b *build.Builder

	BeforeEach(func() {
		// Arrange (shared)
		var err error
		b, err = build.New(registry(),
			[]check.Check{fakeCheck{check.Warning, 5}, fakeCheck{check.Error, 2}},
			fakeExporter{id: "a", supports: true}, fakeExporter{id: "b", supports: false})
		Expect(err).NotTo(HaveOccurred())
	})

	It("builds from the family's defaults with the given changes, and offers the formats that fit", func() {
		// Arrange
		v := param.Values{"size": 7}

		// Act
		r, err := b.Build("box", v, check.Context{})

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Family).To(Equal("box"))
		Expect(r.Values).To(Equal(param.Values{"size": 7}))
		Expect(r.Shape.MinWall).To(BeNumerically("==", 7))
		Expect(r.Findings).To(BeEmpty())
		Expect(r.Formats).To(Equal([]build.Format{{ID: "a", Label: "Format a", Extension: ".a", MediaType: "x/a"}}))
	})

	It("passes the context to the checks", func() {
		// Arrange
		ctx := check.Context{Printer: [3]float64{1, 1, 1}}

		// Act
		r, err := b.Build("box", param.Values{"size": 7}, ctx)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Findings).To(ContainElement(check.Finding{Level: check.Warning, Message: "context seen"}))
	})

	It("clamps values and passes on the notes", func() {
		// Arrange
		v := param.Values{"size": 99}

		// Act
		r, err := b.Build("box", v, check.Context{})

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Values["size"]).To(BeNumerically("==", 20))
		Expect(r.Notes).To(ConsistOf("Size 99 mm is more than the largest allowed, 20 mm; using 20 mm."))
	})

	It("lists every check's findings, and still offers formats for warnings", func() {
		// Arrange
		v := param.Values{"size": 3}

		// Act
		r, err := b.Build("box", v, check.Context{})

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Findings).To(ConsistOf(check.Finding{Level: check.Warning, Message: "warning found"}))
		Expect(r.Formats).To(HaveLen(1))
	})

	It("offers no formats when a check finds an error", func() {
		// Arrange
		v := param.Values{"size": 1}

		// Act
		r, err := b.Build("box", v, check.Context{})

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(r.Findings).To(HaveLen(2))
		Expect(r.Formats).To(BeEmpty())
	})

	It("passes on the family's reason it can't build", func() {
		// Arrange
		v := param.Values{"size": 13}

		// Act
		_, err := b.Build("box", v, check.Context{})

		// Assert
		Expect(err).To(MatchError("Unlucky size. Pick another."))
	})

	It("refuses an unknown family or setting", func() {
		// Arrange
		// (none)

		// Act
		_, errFamily := b.Build("lamp", nil, check.Context{})
		_, errSetting := b.Build("box", param.Values{"spout": 1}, check.Context{})

		// Assert
		Expect(errFamily).To(MatchError(`unknown family "lamp"`))
		Expect(errSetting).To(MatchError(`unknown setting "spout"`))
	})
})

var _ = Describe("build.New", func() {
	It("refuses two exporters with the same id", func() {
		// Arrange
		e := fakeExporter{id: "a"}

		// Act
		_, err := build.New(registry(), nil, e, e)

		// Assert
		Expect(err).To(MatchError(`format "a" is registered twice`))
	})
})

var _ = Describe("Builder.Export", func() {
	var b *build.Builder

	BeforeEach(func() {
		// Arrange (shared)
		var err error
		b, err = build.New(registry(), []check.Check{fakeCheck{check.Error, 2}},
			fakeExporter{id: "a", supports: true}, fakeExporter{id: "b", supports: false},
			fakeExporter{id: "c", supports: true, fail: errors.New("disk full")})
		Expect(err).NotTo(HaveOccurred())
	})

	It("writes the built shape in the chosen format", func() {
		// Arrange
		var buf bytes.Buffer

		// Act
		f, err := b.Export("box", param.Values{"size": 5}, check.Context{}, "a", &buf)

		// Assert
		Expect(err).NotTo(HaveOccurred())
		Expect(f.ID).To(Equal("a"))
		Expect(buf.String()).To(Equal("a data"))
	})

	DescribeTable("refuses what it can't write",
		func(size float64, format, want string) {
			// Arrange
			var buf bytes.Buffer

			// Act
			_, err := b.Export("box", param.Values{"size": size}, check.Context{}, format, &buf)

			// Assert
			Expect(err).To(MatchError(want))
			Expect(buf.Len()).To(BeZero())
		},
		Entry("a shape with an error finding", 1.0, "a", "This shape can't be downloaded: error found"),
		Entry("an unknown format", 5.0, "obj", `unknown format "obj"`),
		Entry("a format that can't hold the shape", 5.0, "b", "Format b isn't available for this shape."),
		Entry("a shape that can't be built", 13.0, "a", "Unlucky size. Pick another."),
	)

	It("passes on a write error", func() {
		// Arrange
		var buf bytes.Buffer

		// Act
		_, err := b.Export("box", param.Values{"size": 5}, check.Context{}, "c", &buf)

		// Assert
		Expect(err).To(MatchError("disk full"))
	})
})
