package api_test

import (
	"github.com/kinglet-dev/saymake/families/vessel"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler: build and export", func() {
	It("builds from slider values", func() {
		// Arrange
		h := newHandler()

		// Act
		r, _ := call(h, `{"op": "build", "family": "tray", "values": {"length": 300, "sides": 4}}`)

		// Assert
		Expect(r.OK).To(BeTrue(), r.Error)
		Expect(r.Reading).To(BeNil())
		Expect(r.Shape.Size[0]).To(BeNumerically("~", 300, 1e-6))
	})

	It("passes on the checks' findings", func() {
		// Arrange
		h := newHandler()

		// Act
		r, _ := call(h, `{"op": "build", "family": "vessel", "values": {"wall": 0.5}}`)

		// Assert
		Expect(r.OK).To(BeTrue(), r.Error)
		Expect(r.Shape.Findings).To(HaveLen(1))
		Expect(r.Shape.Findings[0].Level).To(Equal("warning"))
		Expect(r.Shape.Findings[0].Message).To(HavePrefix("The thinnest wall is about"))
	})

	It("answers slider values that can't be built with the reason and the settings", func() {
		// Arrange
		h := newHandler()

		// Act
		r, out := call(h, `{"op": "build", "family": "vessel", "values": {"wall": 10, "diameter": 20}}`)

		// Assert
		Expect(r.OK).To(BeFalse())
		Expect(r.Error).To(ContainSubstring("The wall is too thick"))
		Expect(r.Shape.Params).To(HaveLen(len(vessel.New().Params())))
		Expect(out.Mesh).To(BeNil())
	})

	It("exports a file named after the object", func() {
		// Arrange
		h := newHandler()

		// Act
		r, out := call(h, `{"op": "export", "family": "vessel", "values": {}, "format": "stl", "name": "bud-vase"}`)

		// Assert
		Expect(r.OK).To(BeTrue(), r.Error)
		Expect(r.File.Name).To(Equal("bud-vase.stl"))
		Expect(r.File.MediaType).To(Equal("model/stl"))
		Expect(len(out.File)).To(BeNumerically(">", 84))
		Expect(out.Mesh).To(BeNil())
	})

	It("names the file \"shape\" when the given name isn't a plain id", func() {
		// Arrange
		h := newHandler()

		// Act
		r, _ := call(h, `{"op": "export", "family": "flat", "values": {}, "format": "svg", "name": "../../evil"}`)

		// Assert
		Expect(r.OK).To(BeTrue(), r.Error)
		Expect(r.File.Name).To(Equal("shape.svg"))
	})

	It("answers an export it can't make with the reason", func() {
		// Arrange
		h := newHandler()

		// Act
		r, out := call(h, `{"op": "export", "family": "flat", "values": {}, "format": "stl"}`)

		// Assert
		Expect(r.OK).To(BeFalse())
		Expect(r.Error).To(Equal("STL isn't available for this shape."))
		Expect(out.File).To(BeNil())
	})
})
