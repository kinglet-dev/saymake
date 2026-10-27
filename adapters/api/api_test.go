package api_test

import (
	"encoding/binary"
	"encoding/json"

	"github.com/kinglet-dev/saymake/adapters/api"
	"github.com/kinglet-dev/saymake/internal/wiring"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type param struct {
	Name, Label, Unit     string
	Min, Max, Step, Value float64
	Whole                 bool
}

type shape struct {
	Family, FamilyLabel string
	Params              []param
	Notes               []string
	Findings            []struct{ Level, Message string }
	Formats             []struct{ ID, Label, Extension string }
	Triangles           int
	Outlines, Scores    [][]float64
	Size                []float64
}

type response struct {
	OK      bool
	Error   string
	Reading *struct {
		Concept, Family string
		Path, Unknown   []string
		Notes           []string
		Uses            []struct{ Typed, Meaning, Kind string }
	}
	Shape *shape
	File  *struct{ Name, MediaType string }
}

func call(h *api.Handler, req string) (response, api.Response) {
	out := h.Handle([]byte(req))
	var r response
	Expect(json.Unmarshal(out.JSON, &r)).To(Succeed(), string(out.JSON))
	return r, out
}

func newHandler() *api.Handler {
	app, err := wiring.New()
	Expect(err).NotTo(HaveOccurred())
	return api.New(app.Interpreter, app.Builder, app.Families)
}

var _ = Describe("Handler: read", func() {
	It("reads a description and builds its shape, with the mesh as binary", func() {
		// Arrange
		h := newHandler()

		// Act
		r, out := call(h, `{"op": "read", "text": "tall fluted vase 20 cm"}`)

		// Assert
		Expect(r.OK).To(BeTrue(), r.Error)
		Expect(r.Reading.Concept).To(Equal("vase"))
		Expect(r.Reading.Path).To(Equal([]string{"Vessel", "Vase"}))
		Expect(r.Reading.Uses).To(ContainElement(struct{ Typed, Meaning, Kind string }{"fluted", "fluted", "feature"}))
		Expect(r.Shape.Family).To(Equal("vessel"))
		Expect(r.Shape.FamilyLabel).To(Equal("Vessel"))
		Expect(r.Shape.Params[0]).To(Equal(param{Name: "height", Label: "Height", Unit: "mm", Min: 20, Max: 400, Step: 1, Value: 200}))
		Expect(r.Shape.Formats).To(HaveLen(2))
		Expect(r.Shape.Size[2]).To(BeNumerically("~", 200, 1e-6))
		nv := binary.LittleEndian.Uint32(out.Mesh[0:4])
		nt := binary.LittleEndian.Uint32(out.Mesh[4+12*nv:])
		Expect(int(nt)).To(Equal(r.Shape.Triangles))
		Expect(out.Mesh).To(HaveLen(8 + 12*int(nv) + 12*int(nt)))
		Expect(out.File).To(BeNil())
	})

	It("sends a flat shape's outlines as x, y pairs, with no mesh", func() {
		// Arrange
		h := newHandler()

		// Act
		r, out := call(h, `{"op": "read", "text": "square plaque 20 cm"}`)

		// Assert
		Expect(r.OK).To(BeTrue(), r.Error)
		Expect(out.Mesh).To(BeNil())
		Expect(r.Shape.Outlines).To(HaveLen(1))
		Expect(r.Shape.Outlines[0]).To(HaveLen(8))
		Expect(r.Shape.Scores).To(HaveLen(1))
		Expect(r.Shape.Formats[0].ID).To(Equal("svg"))
		Expect(r.Shape.Size[0]).To(BeNumerically("~", 200, 1e-6))
	})

	It("answers a description it can't use with the reason", func() {
		// Arrange
		h := newHandler()

		// Act
		r, _ := call(h, `{"op": "read", "text": "a lamp"}`)

		// Assert
		Expect(r.OK).To(BeFalse())
		Expect(r.Error).To(HavePrefix("I couldn't tell what to make."))
	})

	It("answers a reading whose shape can't be built with the reading and the reason", func() {
		// Arrange
		h := newHandler()

		// Act
		r, _ := call(h, `{"op": "read", "text": "bud vase with 8 mm thick walls"}`)

		// Assert
		Expect(r.OK).To(BeFalse())
		Expect(r.Reading.Concept).To(Equal("bud-vase"))
		Expect(r.Error).To(ContainSubstring("The wall is too thick"))
		Expect(r.Shape.Family).To(Equal("vessel"), "the settings come back so they can be fixed")
		Expect(r.Shape.Params).To(ContainElement(param{Name: "wall", Label: "Wall", Unit: "mm", Min: 0.4, Max: 10, Step: 0.1, Value: 8}))
		Expect(r.Shape.Formats).To(BeEmpty())
	})
})
