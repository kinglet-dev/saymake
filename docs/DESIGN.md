# Saymake: design

How Saymake turns a description like "a tall fluted vase, 20 cm" into a
sensible, makeable shape with a hierarchy of concepts, fuzzy word matching
and parametric shape families: no AI model and no network. This covers the
extension points, the algorithms, the architecture and the threat model.

## How it works

```
"tall fluted vase, 20 cm"
   │ words ─▶ tokens ─▶ quantities ("20 cm") and counts ("8 flutes") taken out
   │ remaining words ─▶ fuzzy match (longest phrase first) against the vocabulary
   ▼
interpretation: object = vase (vessel ▸ vase), tall, fluted, height = 200 mm
   │ concept defaults, inherited down the hierarchy, then modifiers, then sizes
   ▼
parameter values (each clamped to its family's range, every change recorded)
   │ family plugin builds the shape
   ▼
mesh (3D) and/or outlines (2D) ─▶ checks (closed mesh, wall, overhang)
   ▼                                 ─▶ exporters: STL, 3MF, SVG
```

Every step is recorded in the interpretation, so the page can show the reading
back ("vase · tall · fluted · height 200 mm") with each value on a slider.

## Extension points (Open/Closed)

| Extension point | Interface | Add one by |
|---|---|---|
| Shape family | `family.Family`: id, parameter specs, `Build(values) (Shape, error)` | a new package under `families/`, registered in the composition root |
| Vocabulary | JSON data: concepts (with parents), modifiers, size words, stop words | editing or adding a vocabulary file; no code |
| Output format | `export.Exporter`: id, extension, media type, `Supports(shape)`, `Write(w, shape)` | a new package under `adapters/export/`, registered in the composition root |
| Shape check | `check.Check`: `Run(shape, context) []Finding` (solid, wall, overhang, bridge, build volume, tipping) | a new check, registered in the composition root |

A new family needs only its package and vocabulary entries that point to its
id and parameter names; the interpreter, page and exporters stay as they are.
The vocabulary loader rejects entries that name an unknown family or
parameter, so a typo is caught at load time instead of silently ignored.

Third-party families would run out of process (JSON over standard input and
output, or WebAssembly), so a plugin can't reach into the core. Not built yet.

## Concept hierarchy

A concept has an id, an optional parent, a family (inherited if absent),
synonyms, and default parameter values (merged down the chain, child wins).
Example: `object ▸ vessel (family vessel) ▸ vase ▸ bud vase`. A word that only
matches a broad concept (e.g. "vessel") still yields a shape.

Words the vocabulary knows (anywhere) match only exactly, never as a typo of
another word; a typo exactly as close to two meanings is left unused with a
note asking which was meant. Unsupported entries are words for things the
tool can't make yet, each with a message shown instead of "not
understood". A phrase may have only one meaning (checked at load).

Modifiers are words or phrases ("tall", "narrow neck", "fluted") with a list of
operations on named parameters: `set`, `scale` or `add`. A modifier may list
the families it applies to; elsewhere it is reported as "not used for a tray"
instead of being dropped silently.

Size words map a dimension word to a parameter per family ("tall" → height,
"wide" → diameter for vessels, width for trays). A quantity without a size
word goes to the family's main dimension.

## Algorithms (established, cited)

- **Fuzzy matching:** optimal string alignment distance (restricted
  Damerau–Levenshtein; Damerau 1964, computed with the Wagner–Fischer dynamic
  program, 1974), with a length-based limit: 0 edits up to 3 letters, 1 up to
  7, 2 beyond. Ties go to fewer edits, then the longer phrase, then vocabulary
  order, so results are deterministic.
- **Phrase matching:** greedy longest match first over word n-grams (the
  maximum-matching method used in dictionary-based tokenization).
- **Vessels:** surface of revolution from a radius profile r(z) built from
  smooth terms (linear taper, sine bulge, neck pinch with a smoothstep
  window), modulated around the axis: flutes r·(1 + d·cos kθ), regular polygon
  sides r·cos(π/n)/cos((θ mod 2π/n) − π/n), and twist θ + τ·z/h.
  Meshed as rings of quads joined by triangle pairs (a structured grid).
  A **slanted rim** cuts the top with a plane tilted by the slant, through
  the top on the handle's side: each column of the grid runs from the base
  to its own rim height h − tan α · R · (1 − cos θ), with R the top radius,
  and the profile is evaluated at each point's true height, so the walls
  keep their shape below the cut (exact for a round, straight top; a smooth
  close fit where the top is patterned or curved). A **lean** is a shear,
  x += z · tan β, applied to the finished solid: the base stays flat, and
  every horizontal slice keeps its shape and area, so the volume is
  unchanged. The overhang becomes atan(tan φ + tan β) for a profile
  slope φ, and the wall across it thins by the cosine of that angle.
- **Trays and flat outlines:** superellipse (Lamé curve, Piet Hein's form)
  x = a·sgn(cos t)|cos t|^(2/n), y = b·sgn(sin t)|sin t|^(2/n), and regular
  polygons. The inner wall is the outline offset inward by the wall
  thickness, using the convex-polygon offset: shift each edge along its
  normal and intersect neighbors (the outlines are convex by construction).
  Floors and bases are fan-triangulated (valid for convex polygons).
- **Mesh checks:** closed, consistently oriented 2-manifold (every edge used
  exactly twice, once in each direction; Botsch et al., *Polygon Mesh
  Processing*, 2010), and volume by the divergence theorem (sum of signed
  tetrahedra; Zhang & Chen 2001), which must be positive.
- **Tipping:** the center of weight (empty) is the volume-weighted mean of
  the signed tetrahedra's centroids; the shape stands on the convex hull
  (Andrew's monotone chain, 1979) of its lowest points. A center outside
  that hull gets a warning that it may tip over.
- **Formats:** binary STL (3D Systems' StereoLithography Interface
  Specification, 1988), 3MF Core Specification 1.3 (3MF Consortium; zip with
  `3D/3dmodel.model`, `[Content_Types].xml`, `_rels/.rels`), SVG 1.1 with a
  millimeter `width`/`height` and matching `viewBox`.

## Architecture

```
domain/        geometry, mesh checks, outlines, parameters, families (interfaces),
               vocabulary model, text (tokens, distance, quantities)
app/           interpret (words → interpretation), build (interpretation → shape + findings)
families/      vessel, tray, flat (plugins)
adapters/      vocabjson (loader with limits, default vocabulary embedded),
               export/{stl,threemf,svg}, api (JSON requests for the page)
cmd/           saymake-wasm (composition root for the browser)
web/           page source (HTML, CSS, JS modules), tests, build script
```

Dependencies point inward. Go standard library only; Ginkgo and Gomega for
tests (MIT). The page's JavaScript has no runtime dependencies; esbuild (MIT)
bundles it at build time, and Playwright (Apache 2.0) and axe-core (MPL 2.0,
used only in CI) test it. None of these ship.

## The page

The build (`web/build.mjs`) makes the same tested page two ways:

- **One self-contained file,** `saymake_<version>.html`: CSS, script, font
  (Outfit, OFL 1.1) and the WebAssembly module are inlined, so it opens by
  double-click with no server and no network. A Content-Security-Policy meta
  tag allows only the inline script and style by hash, WebAssembly
  compilation, and `blob:` downloads; `connect-src` is `'none'`.
- **A kit for another page,** as kinglet.dev embeds it: the tool's markup
  (`_web.html`), its styles (`_web.css`, all under `.saymake`, with classes
  prefixed `sm-` so a host page's rules can't reach them), its bundled
  modules (`_web.js`), the tool (`.wasm`) and Go's runner (`_wasm_exec.js`).
  The module loads the runner and the tool from beside itself. A browser test
  runs the kit in a host page with clashing rules, under the site's policy
  (`script-src 'self' 'wasm-unsafe-eval'; connect-src 'self'`).

- One primary action: type a description, press **Make it**.
- The reading appears as chips (object path, each word used, each size), plus
  words not understood, so a wrong guess is visible and fixable.
- Sliders for every parameter, with units, generated from the family's specs.
- Preview: WebGL with a hand-written orbit (drag, or arrow keys when focused);
  2D outlines drawn as SVG.
- Checks listed in words (e.g. "Wall 0.6 mm is under the 0.8 mm minimum").
- Download buttons for the formats the shape supports.
- Brand tokens, light and dark themes, WCAG 2.2 AA checked with axe-core,
  keyboard-only use, `prefers-reduced-motion`.

## Threat model

- **Trusts:** the compiled code and the browser. Nothing else.
- **Untrusted input:**
  - The description: at most 500 bytes and 60 words; longer input is
    refused with a message. Numbers are parsed with strict rules (no
    exponents, at most 4 digits before the point, 2 after, must be finite).
  - Parameter values from the page: must be finite numbers for known
    parameters; anything else is refused; values outside a parameter's range
    are clamped and the clamp is reported.
  - Vocabulary JSON (embedded today, editable later): at most 256 KB, 400
    concepts, 400 modifiers, 32 phrases per entry, hierarchy depth 12; unknown
    fields, unknown families or parameters, missing parents and cycles are
    refused. Decoded with `encoding/json` strict mode.
  - API requests from the page: JSON at most 16 KB, strict fields; an
    optional printer size of exactly three numbers from 10 to 2000 mm.
- **Resource limits:** mesh resolution is bounded per family (at most 256
  segments around and 256 rings), so a shape has at most about 270,000
  triangles (about 13 MB of STL); every family's worst case is tested.
- **Touches:** no files, no network, no processes. Downloads happen only when
  the user presses a button, as browser `Blob` downloads.
- **Output safety:** SVG output is written by our code from numbers only, with
  no text from the description, so it can carry no script. File names are
  built from vocabulary ids, never from typed text.
- **What could go wrong:** a part of a multi-part mesh (a handle) inside out
  while the whole has volume (each connected part's volume is checked);
  hostile text exhausting memory (limits above);
  NaN or infinity reaching geometry (refused at the boundary, and checked
  again before export); a non-manifold or inside-out mesh reaching a slicer
  (checked before every export, and export is refused if it fails).
- **Fuzzed:** the description parser, the quantity parser, the vocabulary
  loader and the API request decoder (`go test -fuzz`).

## Not supported yet

Negation ("not too tall"), relations between parts ("a vase shaped like a
cat"), lids, spouts, stems, legs, dividers, surface textures, text and
logos, threads and hinges. Words for these are recognized and answered with
"not supported yet" rather than ignored.

## Printing limits

The checks use a 0.8 mm minimum wall (two lines of a 0.4 mm nozzle), warn
past 45° from upright, warn when a flat ceiling facing down (the recess
under a raised foot) is wider than 30 mm, warn when a solid doesn't fit the
printer (256 × 256 × 256 mm unless the page is told otherwise), and warn
when a shape's center of weight isn't over what it stands on. Printers
differ; the bridging limit especially is a conservative guess.
