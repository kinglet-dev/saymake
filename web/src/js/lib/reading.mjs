// Words for the page: how a description was read, settings' values with
// their units, sizes, and labels for screen readers and buttons.

// chips lists how the description was read: the object, then each word's
// effect in the order typed, then words not understood.
export function chips(reading) {
  const object = reading.uses.find((u) => u.kind === "object");
  const out = [{ text: reading.path.join(" › "), kind: "object", title: object ? `The object: “${object.typed}”` : "The object" }];
  for (const u of reading.uses) {
    if (u.kind === "object") continue;
    if (u.kind === "unsupported") {
      out.push({ text: `${u.typed}: not supported yet`, kind: "unsupported", title: u.meaning });
      continue;
    }
    if (u.kind === "unused") {
      out.push({ text: `${u.typed}: not used`, kind: "unused", title: `“${u.typed}” (${u.meaning}) doesn't apply to this object` });
      continue;
    }
    out.push({ text: u.meaning, kind: u.kind, title: `From “${u.typed}”` });
  }
  for (const w of reading.unknown || []) {
    out.push({ text: `${w}: not understood`, kind: "unknown", title: `No meaning found for “${w}”` });
  }
  return out;
}

// formatValue writes a setting's value to its step's precision, with its unit.
export function formatValue(param, v) {
  const decimals = param.whole ? 0 : Math.max(0, Math.ceil(-Math.log10(param.step) - 1e-9));
  const n = String(Number(v.toFixed(decimals)));
  if (param.unit === "") return n;
  if (param.unit === "×" || param.unit === "°") return n + param.unit;
  return `${n} ${param.unit}`;
}

// sizeText writes a shape's extent: x × y × z for a solid, x × y if flat.
export function sizeText(size, solid) {
  const parts = (solid ? size : size.slice(0, 2)).map((v) => String(Math.round(v * 10) / 10));
  return `${parts.join(" × ")} mm`;
}

// previewLabel describes the preview for screen readers.
export function previewLabel(name, size, solid) {
  const what = name.toLowerCase();
  if (solid) return `3D preview of the ${what}, ${sizeText(size, true)}. Drag, or use the arrow keys, to turn it.`;
  return `Outline of the ${what}, ${sizeText(size, false)}. Red lines are cut; blue lines are scored.`;
}

// fileLabel is a download button's text.
export function fileLabel(format) {
  return `Download ${format.label} (${format.extension})`;
}
