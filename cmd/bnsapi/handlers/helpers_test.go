package handlers

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestHexbytes(t *testing.T) {
	a := hexbytes("a hexbyte value")
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("cannot marshal: %s", err)
	}
	var b hexbytes
	if err := json.Unmarshal(raw, &b); err != nil {
		t.Fatalf("cannot unmarshal: %s", err)
	}
	if !bytes.Equal(a, b) {
		t.Fatalf("%q != %q", a, b)
	}
}
