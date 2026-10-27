//go:build js && wasm

// Command saymake-wasm is the composition root for the browser page: it
// assembles the tool and exposes one function to the page,
// saymake.call(request) → {json, mesh, file}, then waits until the page
// calls saymake.close(). It holds no logic of its own.
package main

import (
	"syscall/js"

	"github.com/kinglet-dev/saymake/internal/wiring"
)

func main() {
	h := wiring.Handler(wiring.BuiltIn())
	done := make(chan struct{})
	obj := js.Global().Get("Object").New()
	obj.Set("call", js.FuncOf(func(_ js.Value, args []js.Value) any {
		resp := h.Handle([]byte(args[0].String()))
		out := js.Global().Get("Object").New()
		out.Set("json", string(resp.JSON))
		out.Set("mesh", toJS(resp.Mesh))
		out.Set("file", toJS(resp.File))
		return out
	}))
	obj.Set("close", js.FuncOf(func(js.Value, []js.Value) any {
		close(done)
		return nil
	}))
	js.Global().Set("saymake", obj)
	if ready := js.Global().Get("onSaymakeReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}
	<-done
}

func toJS(b []byte) any {
	if b == nil {
		return nil
	}
	a := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(a, b)
	return a
}
