// The result's panels: how the description was read, notes, checks,
// downloads and settings. DOM only; the words come from lib/reading.mjs.
import { chips, fileLabel, formatValue } from "../lib/reading.mjs";

const item = (doc, text, cls) => {
  const li = doc.createElement("li");
  li.textContent = text;
  if (cls) li.className = `sm-${cls}`;
  return li;
};

export function renderChips(ul, reading) {
  const doc = ul.ownerDocument;
  ul.replaceChildren(...chips(reading).map((c) => {
    const li = item(doc, c.text, c.kind);
    li.title = c.title;
    return li;
  }));
}

export function renderNotes(ul, notes) {
  ul.replaceChildren(...notes.map((n) => item(ul.ownerDocument, n)));
  ul.hidden = notes.length === 0;
}

export function renderFindings(ul, findings) {
  const doc = ul.ownerDocument;
  ul.replaceChildren(...findings.map((f) => {
    const li = item(doc, "", f.level === "error" ? "error-level" : "warning-level");
    const level = doc.createElement("span");
    level.className = "sm-level";
    level.textContent = f.level === "error" ? "Can't print" : "Check";
    li.append(level, f.message);
    return li;
  }));
  ul.hidden = findings.length === 0;
}

export function renderDownloads(div, formats, onPick) {
  const doc = div.ownerDocument;
  div.replaceChildren(...formats.map((f) => {
    const b = doc.createElement("button");
    b.type = "button";
    b.className = "sm-button";
    b.textContent = fileLabel(f);
    b.dataset.format = f.id;
    b.addEventListener("click", () => onPick(f.id));
    return b;
  }));
}

// renderSettings makes a slider per setting and returns a function that
// shows new values (after the tool clamps them) without rebuilding the form.
export function renderSettings(form, params, onChange) {
  const doc = form.ownerDocument;
  const rows = params.map((p) => {
    const id = `setting-${p.name}`;
    const row = doc.createElement("div");
    row.className = "sm-setting";
    const label = doc.createElement("label");
    label.htmlFor = id;
    label.textContent = p.label;
    // Plain text, not <output>: outputs are live regions, and a screen
    // reader would announce every slider in the form as each one moves.
    const out = doc.createElement("span");
    out.className = "sm-value";
    out.id = `${id}-value`;
    const input = doc.createElement("input");
    Object.assign(input, { type: "range", id, min: p.min, max: p.max, step: p.step, value: p.value });
    input.setAttribute("aria-describedby", out.id);
    const show = (v) => {
      out.textContent = formatValue(p, v);
      input.setAttribute("aria-valuetext", out.textContent);
    };
    show(p.value);
    input.addEventListener("input", () => {
      show(Number(input.value));
      onChange(p.name, Number(input.value));
    });
    row.append(label, out, input);
    return { p, input, show, row };
  });
  form.replaceChildren(...rows.map((r) => r.row));
  return (values) => {
    for (const r of rows) {
      r.input.value = values[r.p.name];
      r.show(values[r.p.name]);
    }
  };
}

// download hands the browser a file to save.
export function download(doc, bytes, info) {
  const url = URL.createObjectURL(new Blob([bytes], { type: info.mediaType }));
  const a = doc.createElement("a");
  a.href = url;
  a.download = info.name;
  doc.body.append(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
