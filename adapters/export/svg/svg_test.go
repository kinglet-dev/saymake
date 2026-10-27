package svg_test

import (
	"bytes"
	"encoding/xml"
	"errors"

	"github.com/kinglet-dev/saymake/adapters/export/svg"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type doc struct {
	XMLName xml.Name `xml:"http://www.w3.org/2000/svg svg"`
	Width   string   `xml:"width,attr"`
	Height  string   `xml:"height,attr"`
	ViewBox string   `xml:"viewBox,attr"`
	Groups  []struct {
		ID     string `xml:"id,attr"`
		Stroke string `xml:"stroke,attr"`
		Fill   string `xml:"fill,attr"`
		Paths  []struct {
			D string `xml:"d,attr"`
		} `xml:"path"`
	} `xml:"g"`
}

var rect = geom.Outline{{X: -50, Y: -30}, {X: 50, Y: -30}, {X: 50, Y: 30}, {X: -50, Y: 30}}
var inner = geom.Outline{{X: -40, Y: -20}, {X: 40, Y: -20}, {X: 40, Y: 20.125}, {X: -40, Y: 20}}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func write(s family.Shape) (string, doc) {
	var buf bytes.Buffer
	Expect(svg.Exporter{}.Write(&buf, s)).To(Succeed())
	var d doc
	Expect(xml.Unmarshal(buf.Bytes(), &d)).To(Succeed())
	return buf.String(), d
}

var _ = Describe("SVG exporter", func() {
	It("describes itself and supports flat shapes only", func() {
		// Arrange
		e := svg.Exporter{}

		// Act
		got := []string{e.ID(), e.Label(), e.Extension(), e.MediaType()}

		// Assert
		Expect(got).To(Equal([]string{"svg", "SVG", ".svg", "image/svg+xml"}))
		Expect(e.Supports(family.Shape{Outlines: []geom.Outline{rect}})).To(BeTrue())
		Expect(e.Supports(family.Shape{Mesh: &geom.Mesh{}})).To(BeFalse())
	})

	It("sizes the drawing in millimeters to fit the outlines", func() {
		// Arrange
		s := family.Shape{Outlines: []geom.Outline{rect}}

		// Act
		_, d := write(s)

		// Assert
		Expect(d.Width).To(Equal("100mm"))
		Expect(d.Height).To(Equal("60mm"))
		Expect(d.ViewBox).To(Equal("-50 -30 100 60"))
	})

	It("sizes the drawing to fit every outline, not just the first", func() {
		// Arrange: the score line reaches past the cut line on every side
		cut := geom.Outline{{X: 20, Y: -2}, {X: 30, Y: -2}, {X: 30, Y: 8}}
		score := geom.Outline{{X: -10, Y: -10}, {X: 40, Y: -10}, {X: 40, Y: 20}, {X: -10, Y: 20}}

		// Act
		_, d := write(family.Shape{Outlines: []geom.Outline{cut}, Scores: []geom.Outline{score}})

		// Assert
		Expect(d.ViewBox).To(Equal("-10 -20 50 30"))
	})

	It("draws cut lines and score lines in separate, unfilled groups by color", func() {
		// Arrange
		s := family.Shape{Outlines: []geom.Outline{rect}, Scores: []geom.Outline{inner}}

		// Act
		_, d := write(s)

		// Assert
		Expect(d.Groups).To(HaveLen(2))
		Expect([]string{d.Groups[0].ID, d.Groups[0].Stroke, d.Groups[0].Fill}).To(Equal([]string{"cut", "#ff0000", "none"}))
		Expect([]string{d.Groups[1].ID, d.Groups[1].Stroke, d.Groups[1].Fill}).To(Equal([]string{"score", "#0000ff", "none"}))
	})

	It("flips y, since SVG's y axis points down, and closes each path", func() {
		// Arrange
		s := family.Shape{Outlines: []geom.Outline{rect}, Scores: []geom.Outline{inner}}

		// Act
		_, d := write(s)

		// Assert
		Expect(d.Groups[0].Paths[0].D).To(Equal("M-50 30L50 30L50 -30L-50 -30Z"))
		Expect(d.Groups[1].Paths[0].D).To(Equal("M-40 20L40 20L40 -20.125L-40 -20Z"))
	})

	It("leaves out the score group when there are no score lines", func() {
		// Arrange
		s := family.Shape{Outlines: []geom.Outline{rect}}

		// Act
		_, d := write(s)

		// Assert
		Expect(d.Groups).To(HaveLen(1))
	})

	It("rounds coordinates to a thousandth of a millimeter", func() {
		// Arrange
		tri := geom.Outline{{X: 0, Y: 0}, {X: 1.23456, Y: 0}, {X: 0, Y: 1}}

		// Act
		_, d := write(family.Shape{Outlines: []geom.Outline{tri}})

		// Assert
		Expect(d.Groups[0].Paths[0].D).To(Equal("M0 0L1.235 0L0 -1Z"))
	})

	It("contains no script, style or link to anything outside", func() {
		// Arrange
		s := family.Shape{Outlines: []geom.Outline{rect}, Scores: []geom.Outline{inner}}

		// Act
		raw, _ := write(s)

		// Assert
		for _, bad := range []string{"<script", "<style", "href", "<foreignObject", "<image", "url("} {
			Expect(raw).NotTo(ContainSubstring(bad), bad)
		}
		Expect(raw).NotTo(MatchRegexp(`\son[a-z]+\s*=`), "no event handler attributes")
	})

	It("refuses a shape without outlines, and passes on write errors", func() {
		// Arrange
		var buf bytes.Buffer

		// Act
		errSolid := svg.Exporter{}.Write(&buf, family.Shape{Mesh: &geom.Mesh{}})
		errDisk := svg.Exporter{}.Write(failingWriter{}, family.Shape{Outlines: []geom.Outline{rect}})

		// Assert
		Expect(errSolid).To(MatchError("SVG needs a flat shape"))
		Expect(errDisk).To(MatchError("disk full"))
	})
})
