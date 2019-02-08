#!/bin/bash
go run main.go -single-package all -log-warning=false -output tmp `find ../idl/webapi -type f` && (cd tmp/all && GOOS=js GOARCH=wasm go build )
