all: clean docs/wasm dist/wasm dist/glfw

docs/wasm:
	# This outputs a html with the current distribution of wasm_exec.js
	mkdir -p docs/wasm
	GOOS=js GOARCH=wasm go build -o docs/wasm ./01-examples/...
	go run ./cmd/gendocs

# Local dist
dist/wasm:
	mkdir -p dist/wasm
	GOOS=js GOARCH=wasm go build -o dist/wasm ./...

dist/glfw:
	mkdir -p dist/glfw
	go build -ldflags="-s -w" -o dist/glfw ./...

clean:
	rm -rf docs/wasm docs/wasm.html
	rm -rf dist

.PHONY: clean

