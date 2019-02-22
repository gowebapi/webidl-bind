#!/bin/bash
set -ex
go run main.go -go-build=wasm -spec-status=../gowebapi.github.io/mkdocs/docs/status.md -log-warning=false -output ../ -inside-package github.com/gowebapi `find ../idl/idl ../idl/webapi -type f`
