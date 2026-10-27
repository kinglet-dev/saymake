package threemf_test

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"io"

	"github.com/kinglet-dev/saymake/adapters/export/threemf"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func tetra() *geom.Mesh {
	m := &geom.Mesh{}
	a, b := m.Add(geom.Vec3{}), m.Add(geom.Vec3{X: 10.5})
	c, d := m.Add(geom.Vec3{Y: 10}), m.Add(geom.Vec3{Z: 10})
	m.Tri(a, c, b)
	m.Tri(a, b, d)
	m.Tri(b, c, d)
	m.Tri(c, a, d)
	return m
}

type model struct {
	XMLName  xml.Name `xml:"http://schemas.microsoft.com/3dmanufacturing/core/2015/02 model"`
	Unit     string   `xml:"unit,attr"`
	Vertices []struct {
		X float64 `xml:"x,attr"`
		Y float64 `xml:"y,attr"`
		Z float64 `xml:"z,attr"`
	} `xml:"resources>object>mesh>vertices>vertex"`
	Triangles []struct {
		V1 int `xml:"v1,attr"`
		V2 int `xml:"v2,attr"`
		V3 int `xml:"v3,attr"`
	} `xml:"resources>object>mesh>triangles>triangle"`
	Items []struct {
		ObjectID string `xml:"objectid,attr"`
	} `xml:"build>item"`
}

func files(b []byte) map[string]string {
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	Expect(err).NotTo(HaveOccurred())
	out := map[string]string{}
	for _, f := range z.File {
		r, err := f.Open()
		Expect(err).NotTo(HaveOccurred())
		data, err := io.ReadAll(r)
		Expect(err).NotTo(HaveOccurred())
		out[f.Name] = string(data)
	}
	return out
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

var _ = Describe("3MF exporter", func() {
	It("describes itself and supports solids only", func() {
		// Arrange
		e := threemf.Exporter{}

		// Act
		got := []string{e.ID(), e.Label(), e.Extension(), e.MediaType()}

		// Assert
		Expect(got).To(Equal([]string{"3mf", "3MF", ".3mf", "model/3mf"}))
		Expect(e.Supports(family.Shape{Mesh: tetra()})).To(BeTrue())
		Expect(e.Supports(family.Shape{})).To(BeFalse())
	})

	It("writes the package parts the 3MF core specification requires", func() {
		// Arrange
		var buf bytes.Buffer

		// Act
		err := threemf.Exporter{}.Write(&buf, family.Shape{Mesh: tetra()})

		// Assert
		Expect(err).NotTo(HaveOccurred())
		fs := files(buf.Bytes())
		Expect(fs).To(HaveKey("[Content_Types].xml"))
		Expect(fs["[Content_Types].xml"]).To(ContainSubstring(`Extension="model" ContentType="application/vnd.ms-package.3dmanufacturing-3dmodel+xml"`))
		Expect(fs["_rels/.rels"]).To(ContainSubstring(`Target="/3D/3dmodel.model"`))
		Expect(fs["_rels/.rels"]).To(ContainSubstring(`Type="http://schemas.microsoft.com/3dmanufacturing/2013/01/3dmodel"`))
		Expect(fs).To(HaveKey("3D/3dmodel.model"))
	})

	It("writes the mesh in millimeters, with every vertex and triangle", func() {
		// Arrange
		var buf bytes.Buffer

		// Act
		err := threemf.Exporter{}.Write(&buf, family.Shape{Mesh: tetra()})

		// Assert
		Expect(err).NotTo(HaveOccurred())
		var m model
		Expect(xml.Unmarshal([]byte(files(buf.Bytes())["3D/3dmodel.model"]), &m)).To(Succeed())
		Expect(m.Unit).To(Equal("millimeter"))
		Expect(m.Vertices).To(HaveLen(4))
		Expect(m.Vertices[1].X).To(BeNumerically("~", 10.5, 1e-9))
		Expect(m.Triangles).To(HaveLen(4))
		Expect([]int{m.Triangles[0].V1, m.Triangles[0].V2, m.Triangles[0].V3}).To(Equal([]int{0, 2, 1}))
		Expect(m.Items).To(HaveLen(1))
	})

	It("writes the same bytes every time", func() {
		// Arrange
		var a, b bytes.Buffer

		// Act
		Expect(threemf.Exporter{}.Write(&a, family.Shape{Mesh: tetra()})).To(Succeed())
		Expect(threemf.Exporter{}.Write(&b, family.Shape{Mesh: tetra()})).To(Succeed())

		// Assert
		Expect(a.Bytes()).To(Equal(b.Bytes()))
	})

	It("dates every part with one fixed day, so builds don't depend on the clock", func() {
		// Arrange
		var buf bytes.Buffer

		// Act
		Expect(threemf.Exporter{}.Write(&buf, family.Shape{Mesh: tetra()})).To(Succeed())

		// Assert
		z, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		Expect(err).NotTo(HaveOccurred())
		for _, f := range z.File {
			y, m, d := f.Modified.UTC().Date()
			Expect([]int{y, int(m), d}).To(Equal([]int{2026, 1, 1}), f.Name)
		}
	})

	It("refuses a shape without a solid, and passes on write errors", func() {
		// Arrange
		var buf bytes.Buffer

		// Act
		errFlat := threemf.Exporter{}.Write(&buf, family.Shape{})
		errDisk := threemf.Exporter{}.Write(failingWriter{}, family.Shape{Mesh: tetra()})

		// Assert
		Expect(errFlat).To(MatchError("3MF needs a solid shape"))
		Expect(errDisk).To(MatchError(ContainSubstring("disk full")))
	})
})
