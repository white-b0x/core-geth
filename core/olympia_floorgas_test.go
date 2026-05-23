// Copyright 2026 The core-geth Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

// EIP-7623 floor data gas unit tests.
// FloorDataGas formula: tokens = nz*4 + z; floor = TxGas + tokens*10 = 21000 + tokens*10
// where nz = non-zero byte count, z = zero byte count, TxCostFloorPerToken = 10.

import (
	"bytes"
	"testing"
)

// TestFloorGasEmptyPayload verifies that an empty calldata payload produces the base
// transaction gas (21,000) — no tokens, no floor premium.
func TestFloorGasEmptyPayload(t *testing.T) {
	got, err := FloorDataGas(nil)
	if err != nil {
		t.Fatalf("FloorDataGas(nil): unexpected error: %v", err)
	}
	const want = uint64(21_000)
	if got != want {
		t.Fatalf("FloorDataGas(nil) = %d, want %d", got, want)
	}
	got2, err := FloorDataGas([]byte{})
	if err != nil {
		t.Fatalf("FloorDataGas([]byte{}): unexpected error: %v", err)
	}
	if got2 != want {
		t.Fatalf("FloorDataGas([]byte{}) = %d, want %d", got2, want)
	}
}

// TestFloorGasAllNonzero verifies the formula for 1000 non-zero bytes:
// tokens = 1000*4 + 0 = 4000; floor = 21000 + 4000*10 = 61,000.
func TestFloorGasAllNonzero(t *testing.T) {
	data := bytes.Repeat([]byte{0xff}, 1000)
	got, err := FloorDataGas(data)
	if err != nil {
		t.Fatalf("FloorDataGas(1000 nz): unexpected error: %v", err)
	}
	// tokens = 1000*4 = 4000; floor = 21000 + 40000 = 61000
	const want = uint64(61_000)
	if got != want {
		t.Fatalf("FloorDataGas(1000 non-zero) = %d, want %d", got, want)
	}
}

// TestFloorGasAllZero verifies the formula for 1000 zero bytes:
// tokens = 0*4 + 1000 = 1000; floor = 21000 + 1000*10 = 31,000.
func TestFloorGasAllZero(t *testing.T) {
	data := bytes.Repeat([]byte{0x00}, 1000)
	got, err := FloorDataGas(data)
	if err != nil {
		t.Fatalf("FloorDataGas(1000 zero): unexpected error: %v", err)
	}
	// tokens = 1000; floor = 21000 + 10000 = 31000
	const want = uint64(31_000)
	if got != want {
		t.Fatalf("FloorDataGas(1000 zero) = %d, want %d", got, want)
	}
}

// TestFloorGasMixed verifies the formula for a mixed payload (200 non-zero + 300 zero):
// tokens = 200*4 + 300 = 1100; floor = 21000 + 1100*10 = 32,000.
func TestFloorGasMixed(t *testing.T) {
	nz := bytes.Repeat([]byte{0xab}, 200)
	z := bytes.Repeat([]byte{0x00}, 300)
	data := append(nz, z...)
	got, err := FloorDataGas(data)
	if err != nil {
		t.Fatalf("FloorDataGas(mixed): unexpected error: %v", err)
	}
	// tokens = 200*4 + 300 = 1100; floor = 21000 + 11000 = 32000
	const want = uint64(32_000)
	if got != want {
		t.Fatalf("FloorDataGas(200 nz + 300 z) = %d, want %d", got, want)
	}
}

// TestFloorGasSingleZero verifies the formula for a single zero byte:
// tokens = 1; floor = 21000 + 10 = 21,010.
func TestFloorGasSingleZero(t *testing.T) {
	got, err := FloorDataGas([]byte{0x00})
	if err != nil {
		t.Fatalf("FloorDataGas(1 zero): unexpected error: %v", err)
	}
	const want = uint64(21_010)
	if got != want {
		t.Fatalf("FloorDataGas(1 zero) = %d, want %d", got, want)
	}
}
