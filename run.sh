#!/bin/bash
set -ex
go run main.go -go-build=wasm -go-test=host -cross-ref ../gowebapi.github.io/mkdocs/docs/jscrossref.md -spec-status=../gowebapi.github.io/mkdocs/docs/status.md -log-warning=false -output ../ -inside-package github.com/gowebapi `find ../idl/idl ../idl/webapi -type f`
