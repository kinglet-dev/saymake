// Package api answers the page's requests: JSON in, JSON out, with the mesh
// or a file as raw bytes beside it. Every request is untrusted: it is size-
// limited and strictly decoded, and file names come from a plain id only.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"

	"github.com/kinglet-dev/saymake/app/build"
	"github.com/kinglet-dev/saymake/app/interpret"
	"github.com/kinglet-dev/saymake/domain/check"
	"github.com/kinglet-dev/saymake/domain/family"
	"github.com/kinglet-dev/saymake/domain/param"
)

// Limits on a request (see the threat model).
const (
	MaxRequest = 16 << 10
	MaxValues  = 64
)

var plainID = regexp.MustCompile(`^[a-z0-9-]{1,40}$`)

// Response is the answer: JSON, plus mesh bytes for a built solid or file
// bytes for an export.
type Response struct {
	JSON, Mesh, File []byte
}

// The ports the handler uses.
type (
	reader interface {
		Read(string) (interpret.Interpretation, error)
	}
	builder interface {
		Build(string, param.Values, check.Context) (build.Result, error)
		Export(string, param.Values, check.Context, string, io.Writer) (build.Format, error)
	}
	families interface {
		Get(string) (family.Family, bool)
	}
)

// Handler answers requests.
type Handler struct {
	in      reader
	b       builder
	fam     families
	startup error // set when the tool couldn't start
}

// New prepares a handler.
func New(in reader, b builder, f families) *Handler { return &Handler{in: in, b: b, fam: f} }

type request struct {
	Op      string       `json:"op"`
	Text    string       `json:"text"`
	Family  string       `json:"family"`
	Values  param.Values `json:"values"`
	Format  string       `json:"format"`
	Name    string       `json:"name"`
	Printer []float64    `json:"printer"` // build volume x, y, z in mm; optional
}

type answer struct {
	OK      bool        `json:"ok"`
	Error   string      `json:"error,omitempty"`
	Reading *reading    `json:"reading,omitempty"`
	Shape   *shape      `json:"shape,omitempty"`
	File    *fileAnswer `json:"file,omitempty"`
}

type fileAnswer struct {
	Name      string `json:"name"`
	MediaType string `json:"mediaType"`
}

// Unavailable returns a handler that answers every request with the reason
// the tool couldn't start.
func Unavailable(err error) *Handler {
	return &Handler{startup: fmt.Errorf("Saymake couldn't start: %w. Please report this.", err)}
}

// Handle answers one request.
func (h *Handler) Handle(raw []byte) Response {
	if h.startup != nil {
		return refuse(nil, h.startup)
	}
	req, err := decode(raw)
	if err != nil {
		return refuse(nil, err)
	}
	ctx := check.Context{}
	copy(ctx.Printer[:], req.Printer)
	switch req.Op {
	case "read":
		return h.read(req.Text, ctx)
	case "build":
		return h.build(nil, req.Family, req.Values, ctx)
	case "export":
		return h.export(req, ctx)
	}
	return refuse(nil, fmt.Errorf("unknown request %q", req.Op))
}

func decode(raw []byte) (request, error) {
	var req request
	if len(raw) > MaxRequest {
		return req, fmt.Errorf("the request is larger than %d KB", MaxRequest>>10)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return req, fmt.Errorf("the request isn't valid JSON: %w", err)
	}
	if dec.More() {
		return req, fmt.Errorf("unexpected data after the request")
	}
	if len(req.Values) > MaxValues {
		return req, fmt.Errorf("the request has too many settings")
	}
	if len(req.Printer) > 0 && (len(req.Printer) != 3 || slices.ContainsFunc(req.Printer, func(v float64) bool { return v < 10 || v > 2000 })) {
		return req, fmt.Errorf("the printer size must be three numbers from 10 to 2000 mm")
	}
	return req, nil
}

func (h *Handler) read(text string, ctx check.Context) Response {
	got, err := h.in.Read(text)
	if err != nil {
		return refuse(nil, err)
	}
	return h.build(toReading(got), got.Family, got.Values, ctx)
}

func (h *Handler) build(rd *reading, familyID string, values param.Values, ctx check.Context) Response {
	res, err := h.b.Build(familyID, values, ctx)
	f, known := h.fam.Get(familyID)
	if err != nil && !known {
		return refuse(rd, err)
	}
	if err != nil {
		// The settings come back so the page can show them to be fixed.
		vals, _, verr := f.Params().Apply(f.Params().Defaults(), values)
		if verr != nil {
			return refuse(rd, verr)
		}
		s, _ := toShape(f, build.Result{Values: vals})
		return Response{JSON: marshal(answer{Reading: rd, Shape: s, Error: err.Error()})}
	}
	s, mesh := toShape(f, res)
	return Response{JSON: marshal(answer{OK: true, Reading: rd, Shape: s}), Mesh: mesh}
}

func (h *Handler) export(req request, ctx check.Context) Response {
	var buf bytes.Buffer
	f, err := h.b.Export(req.Family, req.Values, ctx, req.Format, &buf)
	if err != nil {
		return refuse(nil, err)
	}
	name := req.Name
	if !plainID.MatchString(name) {
		name = "shape"
	}
	a := answer{OK: true, File: &fileAnswer{Name: name + f.Extension, MediaType: f.MediaType}}
	return Response{JSON: marshal(a), File: buf.Bytes()}
}

func refuse(rd *reading, err error) Response {
	return Response{JSON: marshal(answer{Reading: rd, Error: err.Error()})}
}

// marshal can't fail: answers hold only strings, finite numbers, slices and
// structs.
func marshal(a answer) []byte {
	b, _ := json.Marshal(a)
	return b
}
