# Changelog

All notable changes to Saymake are listed here. It follows
[Semantic Versioning](https://semver.org/): within a major version, a
description that works keeps working (words are added, not taken away), the
files it makes keep their formats, and the kit's files keep their names.

## 0.1.0 (2026-10-27)

### Added

- Describe an object in plain words and get a parametric shape: 24 objects
  in three families (vessels such as vases, bowls, cups, mugs, jars, planters
  and pitchers; trays, dishes and open boxes; flat coasters, plaques, tags and
  bookmarks) and 46 features ("fluted", "hexagonal", "rounded corners",
  "chamfered rim", "slanted rim", "leaning", "raised base", "thick handle",
  "with drainage" and more), with sizes ("20 cm tall", "4in diameter") and
  counts ("8 flutes").
- Typos are forgiven (one or two letters, by word length), but a word it
  knows is never taken for another; a typo as close to two words is asked
  about, not guessed. Words for things it can't make yet (text, lids, spouts,
  legs, textures and more) say so.
- Every setting on a slider; a 3D preview for solids and an outline preview
  for flat shapes.
- Downloads: binary STL and 3MF for solids, SVG in millimeters for flat
  shapes (cut lines red, score lines dashed blue), offered only for shapes
  that pass the checks.
- Checks: a closed, outward-facing solid (each part on its own), a 0.8 mm
  minimum wall, overhangs past 45°, wide flat ceilings, whether it fits the
  printer (256 × 256 × 256 mm unless set otherwise) and whether it would tip
  over.
- The page as one file to use offline, `saymake_<version>.html`, and as a
  kit another page can embed (`saymake_<version>_web.js`, `_web.css`,
  `_web.html`, with the `.wasm` and Go's `_wasm_exec.js`), as
  [kinglet.dev/saymake](https://kinglet.dev/saymake/) does.
