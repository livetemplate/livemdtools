# <<.Title>> - WASM Source

A custom Tinkerdown data source built with TinyGo and WebAssembly.

## Prerequisites

- [TinyGo](https://tinygo.org/getting-started/install/) installed
- Tinkerdown CLI

## Building

```bash
make build
```

This compiles `source.go` to `source.wasm`.

## Testing

```bash
make test
```

Opens the test app at http://localhost:8080.

## WASM Interface

Your source must export these functions:

### `fetch() uint64`

Fetch data without arguments. Returns pointer+length packed as uint64.

### `fetchWithArgs(argsPtr, argsLen uint32) uint64`

Fetch data with JSON arguments. Returns pointer+length packed as uint64.

### Return Format

Return a JSON array of objects:

```json
[
  {"key": "value", "other": "data"},
  {"key": "value2", "other": "data2"}
]
```

Or return an error:

```json
{"error": "Something went wrong"}
```

## Customizing

1. Edit `source.go` to fetch from your API
2. Run `make build`
3. Test with `make test`
4. Copy `source.wasm` to your Tinkerdown app
