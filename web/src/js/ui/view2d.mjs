// The flat preview: outlines as SVG, cut lines red and score lines dashed
// blue. Built with DOM calls, never from markup strings.
import { toSVG } from "../lib/outline.mjs";

const NS = "http://www.w3.org/2000/svg";

export function showOutlines(container, cuts, scores) {
  const model = toSVG(cuts, scores, 4);
  const svg = document.createElementNS(NS, "svg");
  svg.setAttribute("viewBox", model.viewBox);
  svg.setAttribute("aria-hidden", "true");
  for (const [cls, list] of [["cut", model.cuts], ["score", model.scores]]) {
    for (const d of list) {
      const p = document.createElementNS(NS, "path");
      p.setAttribute("class", `sm-${cls}`);
      p.setAttribute("d", d);
      p.setAttribute("vector-effect", "non-scaling-stroke");
      svg.append(p);
    }
  }
  container.replaceChildren(svg);
}
