package e2e_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The page runs saymake built for WebAssembly. These specs build that
// module with coverage on, run it under Node.js the way the page does
// (e2e/harness.js), and measure the composition root's coverage.
var _ = Describe("The WebAssembly build", Ordered, func() {
	var wasm, node, wasmExec, covDir string

	BeforeAll(func() {
		// Arrange (shared): build the module with coverage, and find Node.js
		// and Go's wasm_exec.js.
		var err error
		node, err = exec.LookPath("node")
		Expect(err).NotTo(HaveOccurred(), "Node.js is needed to run the WebAssembly build; install it and run the tests again")
		goroot, err := exec.Command("go", "env", "GOROOT").Output()
		Expect(err).NotTo(HaveOccurred())
		wasmExec = filepath.Join(strings.TrimSpace(string(goroot)), "lib", "wasm", "wasm_exec.js")
		dir := GinkgoT().TempDir()
		wasm, covDir = filepath.Join(dir, "saymake.wasm"), filepath.Join(dir, "cover")
		Expect(os.Mkdir(covDir, 0o755)).To(Succeed())
		build := exec.Command("go", "build", "-cover", "-coverpkg=github.com/kinglet-dev/saymake/cmd/...", "-o", wasm, "../cmd/saymake-wasm")
		build.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
		out, err := build.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), string(out))
	})

	run := func(stdin string, extra ...string) []map[string]any {
		cmd := exec.Command(node, append([]string{"harness.js", wasmExec, wasm}, extra...)...)
		cmd.Env = append(os.Environ(), "GOCOVERDIR="+covDir)
		cmd.Stdin = strings.NewReader(stdin)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		Expect(err).NotTo(HaveOccurred(), stderr.String())
		var answers []map[string]any
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			var a map[string]any
			Expect(json.Unmarshal([]byte(line), &a)).To(Succeed(), line)
			answers = append(answers, a)
		}
		return answers
	}

	It("reads a description, builds and exports, answering with JSON and bytes", func() {
		// Arrange
		requests := `{"op": "read", "text": "tall fluted vase 20 cm"}
{"op": "export", "family": "flat", "values": {}, "format": "svg", "name": "coaster"}
{"op": "read", "text": "a lamp"}`

		// Act
		got := run(requests)

		// Assert
		Expect(got).To(HaveLen(3))
		Expect(got[0]["json"].(map[string]any)["ok"]).To(BeTrue())
		Expect(got[0]["mesh"]).To(BeNumerically(">", 1000))
		Expect(got[0]["file"]).To(BeNil())
		Expect(got[1]["json"].(map[string]any)["file"].(map[string]any)["name"]).To(Equal("coaster.svg"))
		Expect(got[1]["file"]).To(BeNumerically(">", 100))
		Expect(got[2]["json"].(map[string]any)["ok"]).To(BeFalse())
	})

	It("is ready without a page callback, and exits when closed", func() {
		// Arrange
		// (no requests)

		// Act
		got := run("", "--no-ready")

		// Assert
		Expect(got).To(Equal([]map[string]any{{"ready": "object"}}))
	})

	It("covers the composition root in full", func() {
		// Arrange
		// (coverage written by the runs above)

		// Act
		out, err := exec.Command("go", "tool", "covdata", "percent", "-i", covDir).CombinedOutput()

		// Assert
		Expect(err).NotTo(HaveOccurred(), string(out))
		Expect(string(out)).To(ContainSubstring("cmd/saymake-wasm\t\tcoverage: 100.0% of statements"))
	})
})
