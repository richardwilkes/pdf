#! /usr/bin/env bash
set -eo pipefail

# Downloads the artifacts from the latest github.com/richardwilkes/mupdf release
# and uses them to refresh the vendored headers (include/mupdf) and the
# per-platform static libraries (lib/). Requires the GitHub CLI (gh).

REPO=richardwilkes/mupdf
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! command -v gh >/dev/null 2>&1; then
  echo "error: the GitHub CLI (gh) is required but was not found in PATH" >&2
  exit 1
fi

TAG=$(gh release view --repo "$REPO" --json tagName --jq .tagName)
echo "Latest $REPO release: $TAG"

WORK_DIR=$(mktemp -d)
trap '/bin/rm -rf "$WORK_DIR"' EXIT

echo "Downloading artifacts..."
gh release download "$TAG" --repo "$REPO" --pattern 'libmupdf_*.tar.gz' --dir "$WORK_DIR"

# Extract every tarball into a staging tree. The headers are identical across
# platforms, so the last extraction wins for include/; each tarball contributes
# its own lib/libmupdf_<platform>.a.
STAGE="$WORK_DIR/stage"
mkdir -p "$STAGE"
for tarball in "$WORK_DIR"/libmupdf_*.tar.gz; do
  echo "Extracting $(basename "$tarball")..."
  tar xzf "$tarball" -C "$STAGE"
done

echo "Updating include/mupdf..."
/bin/rm -rf "$SCRIPT_DIR/include/mupdf"
cp -R "$STAGE/include/mupdf" "$SCRIPT_DIR/include/"

echo "Updating lib/..."
cp "$STAGE"/lib/libmupdf_*.a "$SCRIPT_DIR/lib/"

echo "Done. Updated to $TAG."
