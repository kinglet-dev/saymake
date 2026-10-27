package api_test

import (
	"errors"
	"strings"

	"github.com/kinglet-dev/saymake/adapters/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler: requests at the limits", func() {
	It("accepts a request of exactly the largest size, with the most settings", func() {
		// Arrange
		h := newHandler()
		base := `{"op": "build", "family": "vessel", "values": {` + values(api.MaxValues) + `}, "name": "`
		req := base + strings.Repeat("a", api.MaxRequest-len(base)-2) + `"}`

		// Act
		r, _ := call(h, req)

		// Assert
		Expect(len(req)).To(Equal(api.MaxRequest))
		Expect(r.Error).To(HavePrefix("unknown setting"), "past the limits, to the settings check")
	})
})

var _ = Describe("Handler: the printer size", func() {
	It("warns when the shape doesn't fit the printer given", func() {
		// Arrange
		h := newHandler()

		// Act
		r, _ := call(h, `{"op": "build", "family": "vessel", "values": {}, "printer": [100, 100, 100]}`)

		// Assert
		Expect(r.OK).To(BeTrue(), r.Error)
		Expect(r.Shape.Findings).To(HaveLen(1))
		Expect(r.Shape.Findings[0].Message).To(ContainSubstring("doesn't fit a 100 × 100 × 100 mm printer"))
	})

	It("accepts a printer size at the limits", func() {
		// Arrange
		h := newHandler()

		// Act
		r, _ := call(h, `{"op": "build", "family": "vessel", "values": {}, "printer": [10, 2000, 2000]}`)

		// Assert
		Expect(r.OK).To(BeTrue(), r.Error)
	})

	It("reads with the printer given too", func() {
		// Arrange
		h := newHandler()

		// Act
		r, _ := call(h, `{"op": "read", "text": "vase", "printer": [100, 100, 100]}`)

		// Assert
		Expect(r.Shape.Findings).To(HaveLen(1))
	})

	It("still exports a shape that only gets a warning", func() {
		// Arrange
		h := newHandler()

		// Act
		r, out := call(h, `{"op": "export", "family": "vessel", "values": {}, "format": "stl", "printer": [100, 100, 100]}`)

		// Assert
		Expect(r.OK).To(BeTrue(), r.Error)
		Expect(out.File).NotTo(BeEmpty())
	})
})

var _ = Describe("Handler: hostile or broken requests", func() {
	DescribeTable("refuses with a reason and nothing else",
		func(req, want string) {
			// Arrange
			h := newHandler()

			// Act
			r, out := call(h, req)

			// Assert
			Expect(r.OK).To(BeFalse())
			Expect(r.Error).To(ContainSubstring(want))
			Expect(r.Shape).To(BeNil())
			Expect(out.Mesh).To(BeNil())
			Expect(out.File).To(BeNil())
		},
		Entry("not JSON", `{`, "the request isn't valid JSON"),
		Entry("an unknown field", `{"op": "read", "text": "vase", "admin": true}`, "the request isn't valid JSON"),
		Entry("data after the request", `{"op": "read", "text": "vase"} {}`, "unexpected data after the request"),
		Entry("an unknown op", `{"op": "delete"}`, `unknown request "delete"`),
		Entry("too large", `{"op": "read", "text": "`+strings.Repeat("a", api.MaxRequest)+`"}`, "the request is larger than 16 KB"),
		Entry("too many values", `{"op": "build", "family": "vessel", "values": {`+values(api.MaxValues+1)+`}}`, "too many settings"),
		Entry("a printer size of two numbers", `{"op": "build", "family": "vessel", "values": {}, "printer": [100, 100]}`, "the printer size must be three numbers from 10 to 2000 mm"),
		Entry("a printer size too small", `{"op": "build", "family": "vessel", "values": {}, "printer": [100, 9, 100]}`, "the printer size must be three numbers from 10 to 2000 mm"),
		Entry("a printer size too big", `{"op": "build", "family": "vessel", "values": {}, "printer": [100, 100, 2001]}`, "the printer size must be three numbers from 10 to 2000 mm"),
		Entry("an unknown family", `{"op": "build", "family": "lamp", "values": {}}`, `unknown family "lamp"`),
		Entry("an unknown setting", `{"op": "build", "family": "vessel", "values": {"spout": 1}}`, `unknown setting "spout"`),
	)
})

func values(n int) string {
	var parts []string
	for i := range n {
		parts = append(parts, `"v`+strings.Repeat("x", i)+`": 1`)
	}
	return strings.Join(parts, ",")
}

var _ = Describe("Unavailable", func() {
	It("answers every request with the reason the tool couldn't start", func() {
		// Arrange
		h := api.Unavailable(errors.New("the vocabulary is broken"))

		// Act
		r, out := call(h, `{"op": "read", "text": "vase"}`)

		// Assert
		Expect(r.OK).To(BeFalse())
		Expect(r.Error).To(Equal("Saymake couldn't start: the vocabulary is broken. Please report this."))
		Expect(out.Mesh).To(BeNil())
	})
})
