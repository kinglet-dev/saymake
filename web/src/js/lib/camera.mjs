// Camera math for the 3D preview: 4×4 matrices, column-major as WebGL
// expects, with z up (the build plate is the xy plane).

const FOV = Math.PI / 4;
const MAX_PITCH = 1.4; // radians: stop short of straight down or up

// perspective is the standard OpenGL projection (as gluPerspective).
export function perspective(fovy, aspect, near, far) {
  const f = 1 / Math.tan(fovy / 2);
  const nf = 1 / (near - far);
  return new Float32Array([f / aspect, 0, 0, 0, 0, f, 0, 0, 0, 0, (far + near) * nf, -1, 0, 0, 2 * far * near * nf, 0]);
}

// lookAt is the view matrix of a camera at eye looking at target (as gluLookAt).
export function lookAt(eye, target, up) {
  const z = unit(sub(eye, target));
  const x = unit(cross(up, z));
  const y = cross(z, x);
  return new Float32Array([
    x[0], y[0], z[0], 0, x[1], y[1], z[1], 0, x[2], y[2], z[2], 0,
    -dot(x, eye), -dot(y, eye), -dot(z, eye), 1,
  ]);
}

// multiply returns a·b (b is applied first).
export function multiply(a, b) {
  const out = new Float32Array(16);
  for (let c = 0; c < 4; c++) {
    for (let r = 0; r < 4; r++) {
      let s = 0;
      for (let k = 0; k < 4; k++) s += a[k * 4 + r] * b[c * 4 + k];
      out[c * 4 + r] = s;
    }
  }
  return out;
}

// transform applies a matrix to a point.
export function transform(m, p) {
  return [0, 1, 2].map((r) => m[r] * p[0] + m[4 + r] * p[1] + m[8 + r] * p[2] + m[12 + r]);
}

// orbit turns the camera by a drag of dx, dy pixels.
export function orbit(state, dx, dy) {
  const pitch = Math.max(-MAX_PITCH, Math.min(MAX_PITCH, state.pitch + dy * 0.01));
  return { yaw: state.yaw - dx * 0.01, pitch };
}

// frame places the camera so the whole shape (its bounding sphere) fits the
// view from the orbit's direction.
export function frame(min, max, state, aspect) {
  const center = [0, 1, 2].map((a) => (min[a] + max[a]) / 2);
  const radius = Math.max(Math.hypot(max[0] - min[0], max[1] - min[1], max[2] - min[2]) / 2, 1e-3);
  const distance = (radius / Math.sin(FOV / 2)) * 1.15;
  const dir = [Math.cos(state.pitch) * Math.cos(state.yaw), Math.cos(state.pitch) * Math.sin(state.yaw), Math.sin(state.pitch)];
  const eye = center.map((c, a) => c + distance * dir[a]);
  const near = Math.max(distance - radius * 1.2, distance * 0.01);
  const far = distance + radius * 1.2;
  return { view: lookAt(eye, center, [0, 0, 1]), projection: perspective(FOV, aspect, near, far), near, far, eye };
}

const sub = (a, b) => [a[0] - b[0], a[1] - b[1], a[2] - b[2]];
const dot = (a, b) => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
const cross = (a, b) => [a[1] * b[2] - a[2] * b[1], a[2] * b[0] - a[0] * b[2], a[0] * b[1] - a[1] * b[0]];
const unit = (v) => {
  const l = Math.hypot(...v);
  return v.map((x) => x / l);
};
