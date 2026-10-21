package msgpack_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

// nestedArrays returns msgpack bytes for `n` nested single-element arrays
// ending in an empty array: 0x91 (fixarray len 1) * n, then 0x90.
func nestedArrays(n int) []byte {
	b := make([]byte, 0, n+1)
	for i := 0; i < n; i++ {
		b = append(b, 0x91)
	}
	return append(b, 0x90)
}

// Decoding a deeply nested message must fail with an error rather than
// overflowing the goroutine stack (https://github.com/vmihailenco/msgpack
// nesting DoS).
func TestDecodeNestingLimit(t *testing.T) {
	var out interface{}
	err := msgpack.Unmarshal(nestedArrays(1_000_000), &out)
	if err == nil {
		t.Fatal("expected an error decoding deeply nested input, got nil")
	}
	if !strings.Contains(err.Error(), "nesting depth") {
		t.Fatalf("expected a nesting-depth error, got: %v", err)
	}
}

func TestDecodeDefaultDepthOK(t *testing.T) {
	var out interface{}
	if err := msgpack.Unmarshal(nestedArrays(50), &out); err != nil {
		t.Fatalf("50 levels should decode within the default limit: %v", err)
	}
	var over interface{}
	if err := msgpack.Unmarshal(nestedArrays(200), &over); err == nil {
		t.Fatal("200 levels should exceed the default limit")
	}
}

type node struct {
	Child *node
}

func TestDecodeNestingLimitTyped(t *testing.T) {
	// fixmap(1) {"Child": ...} repeated, ending in nil: a self-referential
	// struct goes through the typed decoder path, not the interface one.
	var b []byte
	for i := 0; i < 1_000_000; i++ {
		b = append(b, 0x81, 0xa5, 'C', 'h', 'i', 'l', 'd')
	}
	b = append(b, 0xc0)
	var n node
	err := msgpack.Unmarshal(b, &n)
	if err == nil || !strings.Contains(err.Error(), "nesting depth") {
		t.Fatalf("expected a nesting-depth error, got: %v", err)
	}
}

func TestDecodeSetMaxDepth(t *testing.T) {
	dec := msgpack.NewDecoder(bytes.NewReader(nestedArrays(300)))
	dec.SetMaxDepth(500)
	var out interface{}
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("300 levels should decode with maxDepth 500: %v", err)
	}

	dec2 := msgpack.NewDecoder(bytes.NewReader(nestedArrays(300)))
	dec2.SetMaxDepth(100)
	var over interface{}
	if err := dec2.Decode(&over); err == nil {
		t.Fatal("300 levels should exceed a maxDepth of 100")
	}
}
