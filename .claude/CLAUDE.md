# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A small Go package (`github.com/richardwilkes/pdf`) that wraps [MuPDF](https://mupdf.com)
via cgo to render PDF pages to images and extract text-search hits, links, and the table
of contents. It also handles password-protected documents. The entire public API lives in a
single file, [pdf.go](pdf.go). A runnable demonstration lives in [example/main.go](example/main.go)
(`go run ./example document.pdf [search]`).

Static MuPDF libraries are vendored in [lib/](lib/) for six platforms (macOS, Linux, and
Windows, each amd64 and arm64), so no system MuPDF install is needed. Because the package uses
cgo, a C toolchain and `CGO_ENABLED=1` are required; cross-compiling needs a matching cross C
toolchain. Targets go 1.26.

## Commands

- `./build.sh` — build everything (`go build -v ./...`)
- `./build.sh --all` — build, lint, and run tests with `-race`
- `./build.sh --lint` — install (if needed) and run golangci-lint
- `./build.sh --test` / `--race` — run tests, optionally with the race detector
- `go test -run TestPDF ./...` — run the single test directly
- `./update_from_release.sh` — refresh the vendored MuPDF headers ([include/mupdf](include/mupdf))
  and per-platform static libs ([lib/](lib/)) by downloading the `libmupdf_*.tar.gz` artifacts from
  the latest [richardwilkes/mupdf](https://github.com/richardwilkes/mupdf) GitHub release. Requires
  the GitHub CLI (`gh`). The resulting `lib/*.a` and headers are committed to the repo.

## Architecture

### cgo + MuPDF binding

[pdf.go](pdf.go) opens with a cgo preamble that `#include`s `mupdf/fitz.h` and links the
per-platform static library (`-lmupdf_<os>_<arch>`) from [lib/](lib/). MuPDF reports many
errors through a C-level `fz_try`/`fz_catch` exception mechanism that cgo cannot cross
safely. The preamble therefore defines `wrapped_fz_*` C functions that run the throwing
calls inside `fz_try`/`fz_catch` and return `NULL`/`0` on failure. **Any MuPDF call that can
"throw" must be invoked through such a wrapper, never directly from Go.**

### Document lifecycle and memory

`New(buffer, maxCacheSize)` validates the `%PDF` prefix, creates an `fz_context`, registers the
document handlers, copies the buffer into C memory (`C.CBytes`), and opens it as an in-memory
stream. The `Document` type embeds a pointer to an unexported `document` that owns three C
resources: `ctx`, `doc`, and `data`. These are freed in `release()` in that paired order (doc,
data, ctx). After release, `ctx`/`doc` are nil; every public method first takes `d.lock` and
checks `released()`, returning a zero value or `ErrDocumentReleased` rather than calling into C.

Cleanup is handled two ways: `runtime.AddCleanup` runs `release()` at GC time, and callers
may call `Release()` for immediate reclamation. `document` is embedded by pointer (rather
than by value) so it lives in its own heap allocation, distinct from the `Document` wrapper.
`runtime.AddCleanup` requires that the cleanup arg (`d.document`) not point into the same
allocation as the tracked pointer (`&d`); otherwise the tracked object can never become
unreachable and the cleanup would never run (it panics at registration time). A `sync.Mutex`
on the document serializes all C calls, so methods are safe to call concurrently but execute
one at a time.

### Coordinate scaling

DPI is converted to a scale factor via `dpiToScale` (`dpi/72`, clamped to 10x to guard
against bad EDID data). `RenderPage` renders at a fixed DPI; `RenderPageForSize` computes a
scale to fit within a max width/height. The same scale is applied to search-hit quads, link
rectangles, and TOC x/y positions so all returned coordinates are in rendered-image pixel
space. Rendered output is always `*image.NRGBA` (RGB device colorspace, alpha=1). MuPDF renders
with premultiplied alpha, so `renderPage` runs `unpremultiply` on each non-opaque, non-transparent
pixel to convert back to the straight alpha `image.NRGBA` expects.

`quadToRect` builds the axis-aligned bounding box from all four corners of a search-hit quad (not
just two), so boxes stay correct for rotated or skewed text; min uses `math.Floor`, max uses
`math.Ceil`.

### Authentication

`RequiresAuthentication()` wraps `fz_needs_password`. `Authenticate(password)` wraps
`fz_authenticate_password` and returns an `AuthenticationStatus` byte; a non-zero value means
success, and the `NoAuthenticationRequiredMask` / `UserAuthenticatedMask` / `OwnerAuthenticatedMask`
bit masks describe the detail.

### Hit limits

Search returns at most `maxHits` boxes, further capped by the package-level `OverallMaxHits`
(default 1000) to guard against untrusted input forcing a huge allocation.

### Conventions

- Page numbers are 0-based internally; PDF link URIs of the form `#page=N` are 1-based and
  are decremented when parsed in `loadLinks`.
- All strings coming from MuPDF pass through `sanitizeString`, which strips non-printable/
  control runes and trims whitespace.
- Errors are predefined sentinel `error` values at the top of the file; return those rather
  than constructing new ones.

## Testing notes

The test in [pdf_test.go](pdf_test.go) asserts exact values (page count, TOC count, search-hit
rectangles, link bounds, image stride/bounds) against a committed fixture in
[testfiles/](testfiles/). These exact numbers depend on the bundled MuPDF version, so a MuPDF
upgrade (via `update_from_release.sh`) will likely require updating the expected values in the test.
