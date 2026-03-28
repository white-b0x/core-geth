package eth

import (
	"bytes"
	"testing"
)

func TestMakeExtraDataDefault(t *testing.T) {
	if !bytes.Contains(makeExtraData(nil), []byte("Core-Geth-v1.13.0")) {
		t.Error("missing extra data default client identifier")
	}
}
