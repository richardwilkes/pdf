#! /usr/bin/env bash
set -eo pipefail

cp ../mupdf/dist/lib/libmupdf* lib/
/bin/rm -rf include/mupdf
cp -R ../mupdf/dist/include/mupdf include/
