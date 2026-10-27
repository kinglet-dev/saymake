// The page's composition root: loads the WebAssembly tool, waits until it's
// ready, and starts the page with the real tool and 3D view. The single-file
// page carries Go's runner and the tool inline; otherwise both are fetched
// from next to this module (the development page, and the kit kinglet.dev
// embeds). No logic of its own; the end-to-end browser tests cover it.
import { createBridge } from "./lib/bridge.mjs";
import { startPage } from "./ui/page.mjs";
import { createView } from "./ui/view3d.mjs";

const here = (name) => new URL(name, import.meta.url).href;

async function wasmBytes() {
  const inline = document.getElementById("wasm-data");
  if (inline) return Uint8Array.from(atob(inline.textContent.trim()), (c) => c.charCodeAt(0));
  return new Uint8Array(await (await fetch(here("saymake.wasm"))).arrayBuffer());
}

async function main() {
  if (!globalThis.Go) await import(here("saymake_wasm_exec.js"));
  const go = new Go();
  const ready = new Promise((resolve) => { globalThis.onSaymakeReady = resolve; });
  const { instance } = await WebAssembly.instantiate(await wasmBytes(), go.importObject);
  go.run(instance);
  await ready;
  startPage(document, createBridge((req) => globalThis.saymake.call(req)), createView);
}

main().catch((e) => {
  document.getElementById("status").textContent = `Saymake couldn't start: ${e.message}`;
});
