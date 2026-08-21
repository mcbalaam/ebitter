generate:
	go generate ./...

wasm: generate
	GOOS=js GOARCH=wasm go build -o ebitter.wasm ./cmd/demo
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" .

# Local test: python3 -m http.server 8080

.PHONY: wasm generate
