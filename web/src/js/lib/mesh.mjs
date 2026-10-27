// Decodes the mesh bytes the tool sends for a solid shape (see Go's
// adapters/api: little-endian uint32 vertex count, float32 x, y, z per
// vertex, uint32 triangle count, uint32 corners per triangle) into flat-
// shaded arrays for WebGL: three corners per triangle, each with the
// triangle's unit normal.
export function decodeMesh(bytes) {
  const dv = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  const need = (n) => {
    if (n > bytes.byteLength) throw new Error("the mesh data is cut short");
  };
  need(4);
  const nv = dv.getUint32(0, true);
  need(4 + 12 * nv + 4);
  const nt = dv.getUint32(4 + 12 * nv, true);
  const end = 8 + 12 * nv + 12 * nt;
  need(end);
  if (end !== bytes.byteLength) throw new Error("unexpected data after the mesh");
  const vert = (i) => [0, 1, 2].map((k) => dv.getFloat32(4 + 12 * i + 4 * k, true));
  const positions = new Float32Array(9 * nt);
  const normals = new Float32Array(9 * nt);
  const min = [Infinity, Infinity, Infinity];
  const max = [-Infinity, -Infinity, -Infinity];
  for (let t = 0; t < nt; t++) {
    const corners = [0, 1, 2].map((k) => {
      const i = dv.getUint32(8 + 12 * nv + 12 * t + 4 * k, true);
      if (i >= nv) throw new Error("a triangle names a missing vertex");
      return vert(i);
    });
    const n = normal(corners);
    corners.forEach((c, k) => {
      positions.set(c, 9 * t + 3 * k);
      normals.set(n, 9 * t + 3 * k);
      for (let a = 0; a < 3; a++) {
        min[a] = Math.min(min[a], c[a]);
        max[a] = Math.max(max[a], c[a]);
      }
    });
  }
  return { positions, normals, count: 3 * nt, min, max };
}

function normal([a, b, c]) {
  const u = [b[0] - a[0], b[1] - a[1], b[2] - a[2]];
  const v = [c[0] - a[0], c[1] - a[1], c[2] - a[2]];
  const n = [u[1] * v[2] - u[2] * v[1], u[2] * v[0] - u[0] * v[2], u[0] * v[1] - u[1] * v[0]];
  const len = Math.hypot(...n);
  return len === 0 ? [0, 0, 0] : n.map((x) => x / len);
}
