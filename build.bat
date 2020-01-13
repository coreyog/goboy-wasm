@echo off
cmd /C "set GOOS=js&& set GOARCH=wasm&& go build -o dist/goboy.wasm main.go"