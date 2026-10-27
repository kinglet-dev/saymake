// Runs the WebAssembly build under Node.js the way the page does: loads
// Go's wasm_exec.js, waits for saymake to be ready, sends the requests
// given as JSON lines on stdin, prints one JSON line per answer, then
// closes the module so it exits (and writes coverage, if asked).
// Usage: node harness.js <wasm_exec.js> <saymake.wasm> [--no-ready]
"use strict";
const fs = require("fs");
globalThis.require = require;
globalThis.fs = fs;
globalThis.path = require("path");
globalThis.TextEncoder = require("util").TextEncoder;
globalThis.TextDecoder = require("util").TextDecoder;
globalThis.performance ??= require("perf_hooks").performance;
globalThis.crypto ??= require("crypto");
require(process.argv[2]);

// After saymake.close() the Go program exits, but a timer the Go runtime
// scheduled before that may still fire, and wasm_exec.js then throws "Go
// program has already exited". That is the end of a run we asked for, not a
// failure; anything else still fails the run.
let closing = false;
process.on("uncaughtException", (e) => {
  if (closing && go.exited && e.message === "Go program has already exited") return;
  console.error(e);
  process.exitCode = 1;
});

const requests = fs.readFileSync(0, "utf8").split("\n").filter((l) => l.trim() !== "");
const go = new Go();
go.env = { GOCOVERDIR: process.env.GOCOVERDIR || "", TMPDIR: require("os").tmpdir() };
go.exit = (code) => { process.exitCode = code; };

function answer(r) {
  return { json: JSON.parse(r.json), mesh: r.mesh ? r.mesh.length : null, file: r.file ? r.file.length : null };
}

if (process.argv[4] !== "--no-ready") {
  globalThis.onSaymakeReady = () => {
    for (const req of requests) console.log(JSON.stringify(answer(globalThis.saymake.call(req))));
    closing = true;
    globalThis.saymake.close();
  };
}
WebAssembly.instantiate(fs.readFileSync(process.argv[3]), go.importObject)
  .then((r) => {
    const exited = go.run(r.instance); // settles when the program ends
    if (process.argv[4] === "--no-ready") {
      // Go may hand control back before main has finished starting.
      const poll = setInterval(() => {
        if (!globalThis.saymake) return;
        clearInterval(poll);
        console.log(JSON.stringify({ ready: typeof globalThis.saymake }));
        closing = true;
        globalThis.saymake.close();
      }, 10);
    }
    return exited;
  })
  .catch((e) => { console.error(e); process.exitCode = 1; });
