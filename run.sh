#!/bin/bash
set -ex
go run main.go -go-build=wasm -log-warning=false -output ../ -inside-package github.com/gowebapi `find ../idl/webapi -type f`
