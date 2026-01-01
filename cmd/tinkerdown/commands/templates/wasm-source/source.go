//go:build tinygo.wasm

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"unsafe"
)

// Response from httpbin.org/get
type HTTPBinResponse struct {
	Args    map[string]string `json:"args"`
	Headers map[string]string `json:"headers"`
	Origin  string            `json:"origin"`
	URL     string            `json:"url"`
}

// DataItem returned to Tinkerdown
type DataItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

//export fetch
func fetch() uint64 {
	return fetchWithArgs(0, 0)
}

//export fetchWithArgs
func fetchWithArgs(argsPtr, argsLen uint32) uint64 {
	// Fetch from httpbin.org
	resp, err := http.Get("https://httpbin.org/get")
	if err != nil {
		return encodeError(err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return encodeError(err.Error())
	}

	var httpbinResp HTTPBinResponse
	if err := json.Unmarshal(body, &httpbinResp); err != nil {
		return encodeError(err.Error())
	}

	// Convert to array of key-value pairs
	items := []DataItem{
		{Key: "origin", Value: httpbinResp.Origin},
		{Key: "url", Value: httpbinResp.URL},
	}
	for k, v := range httpbinResp.Headers {
		items = append(items, DataItem{Key: k, Value: v})
	}

	result, _ := json.Marshal(items)
	return encodeResult(result)
}

func encodeResult(data []byte) uint64 {
	ptr := uint32(uintptr(unsafe.Pointer(&data[0])))
	length := uint32(len(data))
	return (uint64(ptr) << 32) | uint64(length)
}

func encodeError(msg string) uint64 {
	errJSON := []byte(`{"error":"` + msg + `"}`)
	return encodeResult(errJSON)
}

func main() {}
