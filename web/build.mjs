// Builds the page, three ways:
//   build/dev/          the page as separate files, for the browser tests;
//                       build/dev/kit/ also holds the kit and a host page
//   build/kit/          the kit kinglet.dev embeds: web.js (the page's
//                       modules, bundled), web.css (the tool's styles),
//                       web.html (the tool's markup), saymake.wasm, and Go's
//                       saymake_wasm_exec.js with its license
//   build/saymake.html  one self-contained file: open it by double-click
// SAYMAKE_VERSION sets the version the page shows (default "dev").
// It compiles the Go tool to WebAssembly (reproducibly: no paths, build ID
// or commit, so the same source gives the same file on any commit), bundles the modules with esbuild (MIT, build-time only), and gives the
// single file a strict Content-Security-Policy that allows only its inline
// scripts and styles, by hash.
import { execFileSync } from "node:child_process";
import { createHash } from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { build } from "esbuild";

const here = path.dirname(new URL(import.meta.url).pathname);
const src = path.join(here, "src");
const out = path.join(here, "build");
const dev = path.join(out, "dev");
const kit = path.join(out, "kit");
const version = process.env.SAYMAKE_VERSION ?? "dev";
if (!/^(dev|\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?)$/.test(version)) throw new Error(`SAYMAKE_VERSION "${version}" isn't a release version`);
const goroot = execFileSync("go", ["env", "GOROOT"], { encoding: "utf8" }).trim();
const wasmExec = fs.readFileSync(path.join(goroot, "lib", "wasm", "wasm_exec.js"), "utf8");
const goLicense = fs.readFileSync(path.join(goroot, "LICENSE"), "utf8");
const read = (f) => fs.readFileSync(path.join(src, f), "utf8");

fs.rmSync(out, { recursive: true, force: true });
for (const d of [path.join(dev, "fonts"), path.join(dev, "js"), kit]) fs.mkdirSync(d, { recursive: true });

// 1. The tool, as WebAssembly.
const wasmPath = path.join(kit, "saymake.wasm");
execFileSync("go", ["build", "-trimpath", "-buildvcs=false", "-ldflags=-s -w -buildid=", "-o", wasmPath, "./cmd/saymake-wasm"], {
  cwd: path.join(here, ".."), env: { ...process.env, GOOS: "js", GOARCH: "wasm", CGO_ENABLED: "0" }, stdio: "inherit",
});

// 2. The kit: the tool's markup is what index.html has between its markers.
const index = read("index.html");
const start = "<!-- saymake:tool -->", end = "<!-- /saymake:tool -->";
const fragment = index.slice(index.indexOf(start) + start.length, index.indexOf(end)).trim() + "\n";
if (!index.includes(start) || !index.includes(end) || !fragment.startsWith('<div class="saymake">')) {
  throw new Error("index.html must hold the tool's markup, a <div class=\"saymake\">, between its saymake:tool markers");
}
const esm = (await build({ entryPoints: [path.join(src, "js", "main.mjs")], bundle: true, format: "esm", target: "es2022", write: false, legalComments: "none" })).outputFiles[0].text;
fs.writeFileSync(path.join(kit, "web.js"), esm);
fs.writeFileSync(path.join(kit, "web.css"), read("tool.css"));
fs.writeFileSync(path.join(kit, "web.html"), fragment);
fs.writeFileSync(path.join(kit, "saymake_wasm_exec.js"), wasmExec);
fs.writeFileSync(path.join(kit, "saymake_wasm_exec_LICENSE.txt"), goLicense);

// 3. The development page: the source as it is, Go's runner and the tool
// beside its modules; and the kit in a host page that stands in for
// kinglet.dev (tests/fixtures/host.html), with the tool's markup put in.
fs.cpSync(path.join(src, "js"), path.join(dev, "js"), { recursive: true });
for (const f of ["index.html", "base.css", "tool.css"]) fs.copyFileSync(path.join(src, f), path.join(dev, f));
fs.copyFileSync(path.join(src, "fonts", "Outfit-Variable.woff2"), path.join(dev, "fonts", "Outfit-Variable.woff2"));
for (const f of ["saymake.wasm", "saymake_wasm_exec.js"]) fs.copyFileSync(path.join(kit, f), path.join(dev, "js", f));
fs.cpSync(kit, path.join(dev, "kit"), { recursive: true });
const fixtures = path.join(here, "tests", "fixtures");
fs.copyFileSync(path.join(fixtures, "host.css"), path.join(dev, "kit", "host.css"));
fs.writeFileSync(path.join(dev, "kit", "host.html"), fs.readFileSync(path.join(fixtures, "host.html"), "utf8").replace(start, () => fragment));

// 4. The single file.
const iife = (await build({ entryPoints: [path.join(src, "js", "main.mjs")], bundle: true, format: "iife", target: "es2022", write: false, legalComments: "none",
  define: { "import.meta.url": '"about:blank"' } })).outputFiles[0].text; // unused: the runner and tool are inline
const font = fs.readFileSync(path.join(src, "fonts", "Outfit-Variable.woff2")).toString("base64");
const css = (read("base.css") + read("tool.css")).replace("url(fonts/Outfit-Variable.woff2)", `url(data:font/woff2;base64,${font})`);
const wasm = fs.readFileSync(wasmPath).toString("base64");
for (const [name, text] of [["bundle", iife], ["wasm_exec.js", wasmExec]]) {
  if (/<\/script/i.test(text)) throw new Error(`${name} contains "</script", which would end its inline script early`);
}
const hash = (s) => `'sha256-${createHash("sha256").update(s).digest("base64")}'`;
const csp = [
  "default-src 'none'",
  `script-src ${hash(wasmExec)} ${hash(iife)} 'wasm-unsafe-eval'`,
  `style-src ${hash(css)}`,
  "font-src data:", "img-src data: blob:", "connect-src 'none'", "base-uri 'none'", "form-action 'none'",
].join("; ");
const notices = ["Go (wasm_exec.js and the compiled runtime): BSD-3-Clause\n" + goLicense,
  "Outfit font: SIL Open Font License 1.1\n" + read(path.join("fonts", "Outfit-OFL.txt"))];
let html = index;
const swap = (from, to) => {
  if (!html.includes(from)) throw new Error(`index.html no longer contains: ${from}`);
  html = html.replace(from, () => to);
};
swap(/content="default-src[^"]*"/.exec(html)[0], `content="${csp}"`);
swap('<link rel="stylesheet" href="base.css">\n<link rel="stylesheet" href="tool.css">', `<style>${css}</style>`);
swap('<script type="module" src="js/main.mjs"></script>',
  `<script>${wasmExec}</script>\n<script id="wasm-data" type="application/octet-stream">${wasm}</script>\n<script>${iife}</script>`);
swap('<span id="version">dev</span>', `<span id="version">${version}</span>`);
html = html.replace("</html>", `</html>\n<!--\nThird-party notices\n\n${notices.join("\n\n").replaceAll("--", "- -")}\n-->\n`);
fs.writeFileSync(path.join(out, "saymake.html"), html);
console.log(`build/saymake.html: ${(html.length / 1048576).toFixed(1)} MB; build/kit: ${fs.readdirSync(kit).length} files`);
