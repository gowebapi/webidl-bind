#!/bin/bash
go run main.go -single-package webapiall -log-warning=false -output ../webapi `find ../idl/webapi -type f` && (cd tmp/all && GOOS=js GOARCH=wasm go build )
