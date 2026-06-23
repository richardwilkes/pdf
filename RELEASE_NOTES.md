# Release Notes — v1.27.2

This release upgrades the vendored MuPDF, adds Windows/Linux arm64 support, hardens the render path against untrusted
input, and improves link handling. 15 commits.

## Highlights

### MuPDF upgraded to v1.27.2

The vendored static libraries and headers were refreshed to MuPDF v1.27.2 (#4). A new `update_from_release.sh` script
automates pulling the per-platform `libmupdf_*.tar.gz` artifacts from the latest `richardwilkes/mupdf` GitHub release.

### New platforms

Static libraries are now vendored for **Linux arm64** and **Windows arm64**, bringing the total to six supported targets
(macOS, Linux, and Windows, each amd64 and arm64).

### Improved link handling

- Internal links are now resolved through MuPDF's `fz_resolve_link` / `fz_page_number_from_location` rather than a
  hand-rolled `#page=N` URI parser, allowing resolution of internal links to named references rather than page numbers.
- `PageLink` now exposes a **`DestPoint`** giving the target destination's coordinates on the destination page (in the
  same top-left / y-down, pixel-scaled space as link rects). It is `0,0` for destinations without explicit coordinates
  (e.g. `/Fit`) and for external links.

## Resource limits (hardening against untrusted input)

New package-level caps guard against out-of-memory from malicious or malformed documents:

- **`OverallMaxHits`** (default 1000) — caps search-hit boxes returned, regardless of the `maxHits` argument. If not
  greater than zero, searching is skipped entirely.
- **`OverallMaxLinks`** (default 1000) — caps links returned per page.
- **`OverallMaxTOCEntries`** (default 1000) — caps table-of-contents entries, counted across the entire nested outline
  tree.
- **`OverallMaxPixels`** (default `math.MaxInt32 / 4`) — caps pixels (width × height) in a rendered image. A render that
  would exceed this is rejected with the new **`ErrImageTooLarge`**. `RenderPageForSize` checks this up front before
  allocating, and both render paths enforce it centrally in `renderPage`.

## Correctness & robustness fixes

- **Fixed a scaling issue on Intel (amd64).**
- **Thread safety:** all C calls are now serialized by a mutex on the document, so methods
  are safe to call concurrently (they execute one at a time).
- **Search-hit boxes** now enclose all four corners of a hit quad, keeping bounding boxes correct for rotated or skewed
  text.
- `wrapped_fz_open_pdf_document_with_stream` now catches any MuPDF exceptions, preventing cgo from crossing a C-level
  `fz_try`/`fz_catch` boundary on open failures.
- Render code was refactored to share a common render block across paths.

## Tooling & project changes

- Targets **Go 1.26** (`go.mod` bumped from 1.25.3).
- Added a **GitHub Actions build workflow** (`.github/workflows/build.yml`).
- Added a runnable **example program** (`example/main.go`): `go run ./example document.pdf [search]`.
- Added `.gitattributes`; replaced `copy_from_mupdf.sh` with `update_from_release.sh`.
- README updated to reflect the current state.
