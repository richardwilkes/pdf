# pdf

[![Go Reference](https://pkg.go.dev/badge/github.com/richardwilkes/pdf.svg)](https://pkg.go.dev/github.com/richardwilkes/pdf)
[![Build](https://github.com/richardwilkes/pdf/actions/workflows/build.yml/badge.svg)](https://github.com/richardwilkes/pdf/actions/workflows/build.yml)

A small Go package that wraps [MuPDF](https://mupdf.com) v1.27.2 via cgo to render PDF pages to images and to extract
text-search hits, links, and the table of contents.

## Features

- Render any page to an `*image.NRGBA`, either at a fixed DPI or scaled to fit a maximum width and height.
- Return the bounding boxes of search-text matches on a rendered page.
- Extract a page's links (both external URIs and internal page references).
- Extract the document's table of contents.
- Handle password-protected documents.

All returned coordinates (search hits, link bounds, TOC positions) are in the pixel space of the rendered image, so they
line up directly with what you draw.

## Usage

Static [MuPDF](https://mupdf.com) libraries are vendored in [lib/](lib/) for the following platforms, so no system
[MuPDF](https://mupdf.com) installation is required:

| OS      | Architectures   |
|---------|-----------------|
| macOS   | amd64, arm64    |
| Linux   | amd64, arm64    |
| Windows | amd64, arm64    |

Because the package uses cgo, a C toolchain must be available and `CGO_ENABLED=1` (the default for native builds).
Cross-compilation requires an appropriate cross C toolchain.

### Building on Windows

The vendored Windows libraries are built with a **UCRT** mingw-w64 toolchain, so the local C toolchain must also be UCRT
mingw-w64. The MSVCRT variant (`mingw64`) and TDM-GCC do **not** work: they fail at link time with `undefined reference
to '__intrinsic_setjmpex'`, because that symbol is emitted only against UCRT mingw-w64 headers and resolvable only by a
UCRT runtime.

Run [setup-windows.ps1](setup-windows.ps1) from a PowerShell prompt to install Git, Go, and the UCRT mingw-w64 toolchain
(MSYS2 `ucrt64`) and configure `PATH`/`CGO_ENABLED`. Then open a fresh Git Bash (or PowerShell) window and run
`./build.sh`. Build from Git Bash or PowerShell, not from the MSYS2 shell.

The vendored headers ([include/mupdf](include/mupdf)) and static libraries ([lib/](lib/)) are refreshed from a [sibling
repo](https://github.com/richardwilkes/mupdf) via `update_from_release.sh`.

## Example

A complete, runnable program lives in [example/main.go](example/main.go). It renders the first page of a PDF to a
PNG and reports the table of contents, search hits, and links:

```sh
go run ./example document.pdf [search]
```
