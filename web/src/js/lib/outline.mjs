// Turns the tool's flat outlines (x, y pairs in mm, y up) into SVG path
// data (y down) and a viewBox that fits every line, with padding.
export function toSVG(cuts, scores, pad) {
  let lo = [Infinity, Infinity], hi = [-Infinity, -Infinity];
  const paths = (list) => (list || []).map((xs) => {
    let d = "";
    for (let i = 0; i < xs.length; i += 2) {
      const x = xs[i], y = -xs[i + 1];
      lo = [Math.min(lo[0], x), Math.min(lo[1], y)];
      hi = [Math.max(hi[0], x), Math.max(hi[1], y)];
      d += `${i === 0 ? "M" : "L"}${num(x)} ${num(y)}`;
    }
    return d + "Z";
  });
  const c = paths(cuts), s = paths(scores);
  if (lo[0] > hi[0]) return { cuts: c, scores: s, viewBox: "0 0 1 1" };
  const box = [lo[0] - pad, lo[1] - pad, hi[0] - lo[0] + 2 * pad, hi[1] - lo[1] + 2 * pad];
  return { cuts: c, scores: s, viewBox: box.map(num).join(" ") };
}

const num = (v) => String(Math.round(v * 1000) / 1000 + 0); // + 0 turns -0 into 0
