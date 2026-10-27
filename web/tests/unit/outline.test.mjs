import { test } from "node:test";
import assert from "node:assert/strict";
import { toSVG } from "../../src/js/lib/outline.mjs";

test("toSVG flips y for SVG and closes each path", () => {
  // Arrange
  const cuts = [[-50, -30, 50, -30, 50, 30, -50, 30]];

  // Act
  const s = toSVG(cuts, [], 0);

  // Assert
  assert.deepEqual(s.cuts, ["M-50 30L50 30L50 -30L-50 -30Z"]);
  assert.deepEqual(s.scores, []);
  assert.equal(s.viewBox, "-50 -30 100 60");
});

test("toSVG fits every line, score lines included, with padding", () => {
  // Arrange
  const cuts = [[0, 0, 10, 0, 10, 10]];
  const scores = [[-5, -5, 20, -5, 20, 20]];

  // Act
  const s = toSVG(cuts, scores, 2);

  // Assert
  assert.equal(s.viewBox, "-7 -22 29 29");
  assert.equal(s.scores.length, 1);
});

test("toSVG handles no outlines at all", () => {
  // Arrange
  const none = [];

  // Act
  const s = toSVG(none, null, 1);

  // Assert
  assert.equal(s.viewBox, "0 0 1 1");
});
