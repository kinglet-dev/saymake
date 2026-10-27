// Package threemf writes 3MF (3MF Consortium, Core Specification 1.3): a
// zip package with a content-types part, a relationship to the 3D model
// part, and the model itself, in millimeters.
package threemf

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/geom"
)

const contentTypes = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="model" ContentType="application/vnd.ms-package.3dmanufacturing-3dmodel+xml"/>
</Types>
`

const rels = `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Target="/3D/3dmodel.model" Id="rel0" Type="http://schemas.microsoft.com/3dmanufacturing/2013/01/3dmodel"/>
</Relationships>
`

// stamp is a fixed time so the same shape always gives the same file.
var stamp = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// Exporter writes 3MF.
type Exporter struct{}

// ID is the format's id.
func (Exporter) ID() string { return "3mf" }

// Label is the format's name.
func (Exporter) Label() string { return "3MF" }

// Extension is the file extension.
func (Exporter) Extension() string { return ".3mf" }

// MediaType is the file's media type.
func (Exporter) MediaType() string { return "model/3mf" }

// Supports reports whether the shape can be written.
func (Exporter) Supports(s family.Shape) bool { return s.Mesh != nil }

// Write writes the shape's solid.
func (Exporter) Write(w io.Writer, s family.Shape) error {
	if s.Mesh == nil {
		return errors.New("3MF needs a solid shape")
	}
	// The package is built in memory, where writes can't fail, then
	// written out in one go.
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	for _, part := range []struct{ name, body string }{
		{"[Content_Types].xml", contentTypes},
		{"_rels/.rels", rels},
		{"3D/3dmodel.model", model(s.Mesh)},
	} {
		f, _ := z.CreateHeader(&zip.FileHeader{Name: part.name, Method: zip.Deflate, Modified: stamp})
		io.WriteString(f, part.body)
	}
	z.Close()
	_, err := w.Write(buf.Bytes())
	return err
}

func model(m *geom.Mesh) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<model unit="millimeter" xml:lang="en-US" xmlns="http://schemas.microsoft.com/3dmanufacturing/core/2015/02">` + "\n")
	b.WriteString("<resources><object id=\"1\" type=\"model\"><mesh><vertices>\n")
	for _, v := range m.Vertices {
		b.WriteString(`<vertex x="` + num(v.X) + `" y="` + num(v.Y) + `" z="` + num(v.Z) + "\"/>\n")
	}
	b.WriteString("</vertices><triangles>\n")
	for _, t := range m.Triangles {
		b.WriteString(`<triangle v1="` + strconv.Itoa(t[0]) + `" v2="` + strconv.Itoa(t[1]) + `" v3="` + strconv.Itoa(t[2]) + "\"/>\n")
	}
	b.WriteString("</triangles></mesh></object></resources>\n<build><item objectid=\"1\"/></build>\n</model>\n")
	return b.String()
}

func num(f float64) string { return strconv.FormatFloat(f, 'f', -1, 32) }
