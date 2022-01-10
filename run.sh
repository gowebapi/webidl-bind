#!/bin/bash
set -ex
TOP_DIR="../"  # Where repositories should be cloned.
WEBAPI_DIR="${TOP_DIR}/webapi"  # github.com/gowebapi/webapi clone.
IDL_DIR="${TOP_DIR}/idl"  # github.com/gowebapi/idl clone.
DOCS_DIR="${TOP_DIR}/gowebapi.github.io"  # github.com/gowebapi/gowebapi.github.io clone.
go run main.go \
	-go-build=wasm -go-test=host \
	-cross-ref="${DOCS_DIR}/mkdocs/docs/jscrossref.md" \
	-spec-status="${DOCS_DIR}/mkdocs/docs/status.md" \
	-log-warning=false \
	-output "${TOP_DIR}"/ \
	-inside-package github.com/gowebapi \
	$(find ../idl/idl ../idl/webapi -type f)
