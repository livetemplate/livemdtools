.PHONY: build test clean

build:
	tinygo build -o source.wasm -target=wasi -no-debug source.go

test: build
	cd test-app && tinkerdown serve

clean:
	rm -f source.wasm
