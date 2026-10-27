// The 3D preview: a WebGL 2 view of the mesh, turned by dragging or with
// the arrow keys when focused. Flat shading with two lights, drawn on white.
import { frame, multiply, orbit } from "../lib/camera.mjs";

const VERTEX = `#version 300 es
in vec3 position;
in vec3 normal;
uniform mat4 viewProjection;
out vec3 vNormal;
void main() {
  vNormal = normal;
  gl_Position = viewProjection * vec4(position, 1.0);
}`;

const FRAGMENT = `#version 300 es
precision mediump float;
in vec3 vNormal;
uniform vec3 eye;
out vec4 color;
void main() {
  vec3 n = normalize(vNormal);
  float key = max(dot(n, normalize(eye)), 0.0);
  float fill = max(dot(n, normalize(vec3(-0.4, 0.3, 1.0))), 0.0);
  vec3 clay = vec3(0.86, 0.82, 0.74);
  color = vec4(clay * (0.28 + 0.55 * key + 0.25 * fill), 1.0);
}`;

const HOME = { yaw: -0.9, pitch: 0.45 };
const KEY_STEP = 15; // pixels of drag per arrow key press

export function createView(canvas) {
  const gl = canvas.getContext("webgl2", { antialias: true });
  if (!gl) throw new Error("This browser can't show 3D previews (WebGL 2 is off or missing). The downloads still work.");
  const program = link(gl);
  const buffers = { position: gl.createBuffer(), normal: gl.createBuffer() };
  let state = { ...HOME };
  let mesh = null;

  function draw() {
    const scale = globalThis.devicePixelRatio;
    canvas.width = Math.max(1, Math.round(canvas.clientWidth * scale));
    canvas.height = Math.max(1, Math.round(canvas.clientHeight * scale));
    gl.viewport(0, 0, canvas.width, canvas.height);
    gl.clearColor(1, 1, 1, 1);
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
    canvas.dataset.turn = `${state.yaw.toFixed(3)} ${state.pitch.toFixed(3)}`;
    const f = frame(mesh.min, mesh.max, state, canvas.width / canvas.height);
    gl.useProgram(program);
    gl.uniformMatrix4fv(gl.getUniformLocation(program, "viewProjection"), false, multiply(f.projection, f.view));
    gl.uniform3fv(gl.getUniformLocation(program, "eye"), f.eye.map((c, a) => c - (mesh.min[a] + mesh.max[a]) / 2));
    gl.enable(gl.DEPTH_TEST);
    gl.drawArrays(gl.TRIANGLES, 0, mesh.count);
  }

  function attach(name, data) {
    gl.bindBuffer(gl.ARRAY_BUFFER, buffers[name]);
    gl.bufferData(gl.ARRAY_BUFFER, data, gl.STATIC_DRAW);
    const at = gl.getAttribLocation(program, name);
    gl.enableVertexAttribArray(at);
    gl.vertexAttribPointer(at, 3, gl.FLOAT, false, 0, 0);
  }

  let last = null;
  canvas.addEventListener("pointerdown", (e) => {
    last = [e.clientX, e.clientY];
    canvas.setPointerCapture(e.pointerId);
  });
  canvas.addEventListener("pointermove", (e) => {
    if (!last) return;
    state = orbit(state, e.clientX - last[0], e.clientY - last[1]);
    last = [e.clientX, e.clientY];
    draw();
  });
  const release = () => { last = null; };
  canvas.addEventListener("pointerup", release);
  canvas.addEventListener("pointercancel", release);
  canvas.addEventListener("keydown", (e) => {
    const turn = { ArrowLeft: [-KEY_STEP, 0], ArrowRight: [KEY_STEP, 0], ArrowUp: [0, -KEY_STEP], ArrowDown: [0, KEY_STEP] }[e.key];
    if (e.key === "Home") state = { ...HOME };
    else if (turn) state = orbit(state, ...turn);
    else return;
    e.preventDefault();
    draw();
  });
  return {
    show(m) {
      const first = mesh === null;
      mesh = m;
      attach("position", m.positions);
      attach("normal", m.normals);
      draw();
      if (first) new ResizeObserver(draw).observe(canvas); // redraws only once there is a mesh
    },
  };
}

function link(gl) {
  const program = gl.createProgram();
  for (const [type, source] of [[gl.VERTEX_SHADER, VERTEX], [gl.FRAGMENT_SHADER, FRAGMENT]]) {
    const s = gl.createShader(type);
    gl.shaderSource(s, source);
    gl.compileShader(s);
    gl.attachShader(program, s);
  }
  gl.linkProgram(program);
  return program;
}
