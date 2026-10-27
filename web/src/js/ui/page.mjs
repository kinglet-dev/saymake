// The page's flow: read a description, show the reading and the shape,
// rebuild as settings change, and download files. The tool, the 3D view and
// the timers are injected.
import { decodeMesh } from "../lib/mesh.mjs";
import { previewLabel, sizeText } from "../lib/reading.mjs";
import { latest } from "../lib/schedule.mjs";
import { download, renderChips, renderDownloads, renderFindings, renderNotes, renderSettings } from "./panels.mjs";
import { showOutlines } from "./view2d.mjs";

export function startPage(doc, bridge, makeView, timers = globalThis) {
  const $ = (id) => doc.getElementById(id);
  const [text, make, status, errorBox, result] = ["text", "make", "status", "error", "result"].map($);
  let view = null, current = null, showValues = null;

  const say = (msg) => { status.textContent = msg; };
  const printer = () => ["printer-x", "printer-y", "printer-z"].map((id) => Number($(id).value));
  const fail = (msg) => {
    errorBox.textContent = msg;
    errorBox.hidden = !msg;
  };

  function preview(ans) {
    const s = ans.shape;
    const canvas = $("view3d"), flat = $("view2d"), no3d = $("no3d");
    const solid = ans.mesh != null;
    flat.hidden = solid;
    canvas.hidden = no3d.hidden = true;
    if (solid) {
      try {
        view ??= makeView(canvas);
        canvas.hidden = false;
        view.show(decodeMesh(ans.mesh));
        canvas.setAttribute("aria-label", previewLabel(current.title, s.size, true));
      } catch (e) {
        canvas.hidden = true;
        no3d.hidden = false;
        no3d.textContent = e.message;
      }
    } else {
      showOutlines(flat, s.outlines, s.scores);
      flat.setAttribute("role", "img");
      flat.setAttribute("aria-label", previewLabel(current.title, s.size, false));
    }
    $("size").textContent = sizeText(s.size, solid);
  }

  function showShape(ans, notes) {
    fail(ans.error || "");
    renderNotes($("notes"), notes);
    if (!ans.ok) {
      renderFindings($("findings"), []);
      renderDownloads($("downloads"), [], pick);
      $("size").textContent = "Not updated: these settings can't be made.";
      return false;
    }
    renderFindings($("findings"), ans.shape.findings || []);
    renderDownloads($("downloads"), ans.shape.formats || [], pick);
    preview(ans);
    return true;
  }

  function read() {
    const words = text.value.trim();
    if (!words) {
      fail("Type a description first, like “a tall fluted vase, 20 cm”.");
      return;
    }
    const ans = bridge.read(words, printer());
    if (!ans.reading) {
      fail(ans.error);
      result.hidden = true;
      say("");
      return;
    }
    const r = ans.reading;
    current = { family: ans.shape.family, values: {}, name: r.concept, title: r.path[r.path.length - 1] };
    for (const p of ans.shape.params) current.values[p.name] = p.value;
    $("result-title").textContent = current.title;
    renderChips($("reading"), r);
    showValues = renderSettings($("settings"), ans.shape.params, change);
    result.hidden = false;
    const ok = showShape(ans, [...(r.notes || []), ...(ans.shape.notes || [])]);
    say(ok ? `Made a ${current.title.toLowerCase()}, ${sizeText(ans.shape.size, ans.mesh != null)}.` : "Read your description, but couldn't make the shape; see the message below.");
  }

  const rebuild = latest(() => {
    const ans = bridge.build(current.family, current.values, printer());
    if (ans.shape) {
      for (const p of ans.shape.params) current.values[p.name] = p.value;
      showValues(current.values);
    }
    const ok = showShape(ans, (ans.shape && ans.shape.notes) || []);
    say(ok ? `Updated: ${sizeText(ans.shape.size, ans.mesh != null)}.` : "Couldn't make that; see the message below.");
  }, 150, timers);

  function change(name, value) {
    current.values[name] = value;
    rebuild();
  }

  function pick(format) {
    const ans = bridge.exportFile(current.family, current.values, format, current.name, printer());
    if (!ans.ok) {
      fail(ans.error);
      return;
    }
    download(doc, ans.file, ans.fileInfo);
    say(`Downloading ${ans.fileInfo.name}.`);
  }

  $("printer").addEventListener("input", () => {
    if (current) rebuild();
  });
  $("describe").addEventListener("submit", (e) => {
    e.preventDefault();
    read();
  });
  text.disabled = make.disabled = false;
  say("Ready. Describe an object to start.");
  text.focus();
}
