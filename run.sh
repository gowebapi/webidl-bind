#!/bin/bash
set -ex
go run main.go -single-package webapi -log-warning=false -output ../ `find ../idl/webapi -type f`
(cd ../webapi && GOOS=js GOARCH=wasm go build )
