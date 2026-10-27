import { test } from "node:test";
import assert from "node:assert/strict";
import { decodeMesh } from "../../src/js/lib/mesh.mjs";

// encode builds the API's mesh bytes: uint32 vertex count, float32 x, y, z
// per vertex, uint32 triangle count, uint32 corners per triangle.
function encode(vertices, triangles) {
  const buf = new ArrayBuffer(8 + 12 * vertices.length + 12 * triangles.length);
  const dv = new DataView(buf);
  let o = 0;
  dv.setUint32(o, vertices.length, true); o += 4;
  for (const v of vertices) for (const c of v) { dv.setFloat32(o, c, true); o += 4; }
  dv.setUint32(o, triangles.length, true); o += 4;
  for (const t of triangles) for (const i of t) { dv.setUint32(o, i, true); o += 4; }
  return new Uint8Array(buf);
}

const tetra = encode([[0, 0, 0], [10, 0, 0], [0, 10, 0], [0, 0, 10]], [[0, 2, 1], [0, 1, 3], [1, 2, 3], [2, 0, 3]]);

test("decodeMesh unrolls triangles into flat-shaded corners with face normals", () => {
  // Arrange
  const bytes = tetra;

  // Act
  const m = decodeMesh(bytes);

  // Assert
  assert.equal(m.count, 12);
  assert.deepEqual([...m.positions.slice(0, 9)], [0, 0, 0, 0, 10, 0, 10, 0, 0]);
  assert.deepEqual([...m.normals.slice(0, 9)], [0, 0, -1, 0, 0, -1, 0, 0, -1]);
  const k = 1 / Math.sqrt(3);
  for (const c of m.normals.slice(18, 21)) assert.ok(Math.abs(c - k) < 1e-6);
});

test("decodeMesh measures the bounds", () => {
  // Arrange
  const bytes = tetra;

  // Act
  const m = decodeMesh(bytes);

  // Assert
  assert.deepEqual(m.min, [0, 0, 0]);
  assert.deepEqual(m.max, [10, 10, 10]);
});

test("decodeMesh reads data at an offset inside a larger buffer", () => {
  // Arrange
  const big = new Uint8Array(tetra.length + 8);
  big.set(tetra, 4);
  const view = big.subarray(4, 4 + tetra.length);

  // Act
  const m = decodeMesh(view);

  // Assert
  assert.equal(m.count, 12);
});

test("decodeMesh gives a zero normal to a triangle with no area", () => {
  // Arrange
  const bytes = encode([[0, 0, 0], [1, 0, 0], [2, 0, 0]], [[0, 1, 2]]);

  // Act
  const m = decodeMesh(bytes);

  // Assert
  assert.deepEqual([...m.normals.slice(0, 3)], [0, 0, 0]);
});

for (const [name, bytes, message] of [
  ["too short for its counts", tetra.subarray(0, 20), /cut short/],
  ["empty", new Uint8Array(0), /cut short/],
  ["a corner past the vertices", encode([[0, 0, 0]], [[0, 0, 5]]), /missing vertex/],
  ["data left over", new Uint8Array([...tetra, 0, 0, 0, 0]), /unexpected data/],
]) {
  test(`decodeMesh refuses mesh data that is ${name}`, () => {
    // Arrange
    const b = bytes;

    // Act
    const act = () => decodeMesh(b);

    // Assert
    assert.throws(act, message);
  });
}
