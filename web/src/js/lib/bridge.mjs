// The page's side of the WebAssembly module: builds request JSON for
// saymake.call and parses its answers. call is injected (in the page,
// globalThis.saymake.call), so this module has no globals.
export function createBridge(call) {
  const send = (req) => {
    const r = call(JSON.stringify(req));
    let a;
    try {
      a = JSON.parse(r.json);
    } catch {
      return { ok: false, error: "The page couldn't understand the tool's answer. Please reload it and try again." };
    }
    return { ...a, fileInfo: a.file, mesh: r.mesh, file: r.file };
  };
  return {
    read: (text, printer) => send({ op: "read", text, printer }),
    build: (family, values, printer) => send({ op: "build", family, values, printer }),
    exportFile: (family, values, format, name, printer) => send({ op: "export", family, values, format, name, printer }),
  };
}
