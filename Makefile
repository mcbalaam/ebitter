wasm:
	GOOS=js GOARCH=wasm go build -o ebitter.wasm .
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" .

# Local test: python3 -m http.server 8080

.PHONY: wasm
