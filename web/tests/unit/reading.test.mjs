import { test } from "node:test";
import assert from "node:assert/strict";
import { chips, formatValue, sizeText, previewLabel, fileLabel } from "../../src/js/lib/reading.mjs";

test("chips show the object's path, each word's effect, and words not understood", () => {
  // Arrange
  const reading = {
    path: ["Vessel", "Vase"],
    uses: [
      { typed: "tall", meaning: "tall", kind: "feature" },
      { typed: "vase", meaning: "Vase", kind: "object" },
      { typed: "20 cm", meaning: "Height 200 mm", kind: "size" },
      { typed: "8 flutes", meaning: "Flutes 8", kind: "count" },
      { typed: "ribbed", meaning: "fluted", kind: "unused" },
      { typed: "lid", meaning: "Lids aren't supported yet.", kind: "unsupported" },
    ],
    unknown: ["glitter"],
  };

  // Act
  const got = chips(reading);

  // Assert
  assert.deepEqual(got, [
    { text: "Vessel › Vase", kind: "object", title: "The object: “vase”" },
    { text: "tall", kind: "feature", title: "From “tall”" },
    { text: "Height 200 mm", kind: "size", title: "From “20 cm”" },
    { text: "Flutes 8", kind: "count", title: "From “8 flutes”" },
    { text: "ribbed: not used", kind: "unused", title: "“ribbed” (fluted) doesn't apply to this object" },
    { text: "lid: not supported yet", kind: "unsupported", title: "Lids aren't supported yet." },
    { text: "glitter: not understood", kind: "unknown", title: "No meaning found for “glitter”" },
  ]);
});

test("chips say when a feature has a different name than typed", () => {
  // Arrange
  const reading = { path: ["Tray"], uses: [{ typed: "ribbed", meaning: "fluted", kind: "feature" }], unknown: null };

  // Act
  const got = chips(reading);

  // Assert
  assert.deepEqual(got[1], { text: "fluted", kind: "feature", title: "From “ribbed”" });
  assert.equal(got[0].title, "The object");
});

for (const [p, v, want] of [
  [{ unit: "mm", step: 1 }, 200, "200 mm"],
  [{ unit: "mm", step: 0.1 }, 2.25, "2.3 mm"],
  [{ unit: "mm", step: 0.5 }, 3, "3 mm"],
  [{ unit: "×", step: 0.01 }, 0.7499999, "0.75×"],
  [{ unit: "°", step: 1 }, 120, "120°"],
  [{ unit: "", step: 1, whole: true }, 8, "8"],
  [{ unit: "", step: 0.1 }, 5, "5"],
]) {
  test(`formatValue writes ${v} as ${want}`, () => {
    // Arrange
    const param = p;

    // Act
    const got = formatValue(param, v);

    // Assert
    assert.equal(got, want);
  });
}

test("sizeText writes a solid's size as width × depth × height", () => {
  // Arrange
  const size = [110.04, 109.96, 200];

  // Act
  const got = sizeText(size, true);

  // Assert
  assert.equal(got, "110 × 110 × 200 mm");
});

test("sizeText writes a flat shape's size in two dimensions", () => {
  // Arrange
  const size = [90.5, 60, 0];

  // Act
  const got = sizeText(size, false);

  // Assert
  assert.equal(got, "90.5 × 60 mm");
});

test("previewLabel describes the preview for screen readers", () => {
  // Arrange
  const name = "Vase";

  // Act
  const solid = previewLabel(name, [110, 110, 200], true);
  const flat = previewLabel(name, [90, 60, 0], false);

  // Assert
  assert.equal(solid, "3D preview of the vase, 110 × 110 × 200 mm. Drag, or use the arrow keys, to turn it.");
  assert.equal(flat, "Outline of the vase, 90 × 60 mm. Red lines are cut; blue lines are scored.");
});

test("fileLabel names the download", () => {
  // Arrange
  const f = { label: "STL", extension: ".stl" };

  // Act
  const got = fileLabel(f);

  // Assert
  assert.equal(got, "Download STL (.stl)");
});
