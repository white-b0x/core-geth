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

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/params/vars"
)

// TestFloorDataGas tests the EIP-7623 floor data gas calculation.
func TestFloorDataGas(t *testing.T) {
	// Formula: floorDataGas = TxGas + (nz*TxTokenPerNonZeroByte + z) * TxCostFloorPerToken
	// Where: TxGas=21000, TxTokenPerNonZeroByte=4, TxCostFloorPerToken=10
	tests := []struct {
		name    string
		data    []byte
		want    uint64
		wantErr bool
	}{
		{
			name: "empty data",
			data: []byte{},
			want: vars.TxGas, // 21000
		},
		{
			name: "nil data",
			data: nil,
			want: vars.TxGas, // 21000
		},
		{
			name: "single zero byte",
			data: []byte{0x00},
			// tokens = 0*4 + 1 = 1; floor = 21000 + 1*10 = 21010
			want: 21010,
		},
		{
			name: "single nonzero byte",
			data: []byte{0xff},
			// tokens = 1*4 + 0 = 4; floor = 21000 + 4*10 = 21040
			want: 21040,
		},
		{
			name: "10 zero bytes",
			data: make([]byte, 10),
			// tokens = 0*4 + 10 = 10; floor = 21000 + 10*10 = 21100
			want: 21100,
		},
		{
			name: "10 nonzero bytes",
			data: bytes.Repeat([]byte{0x01}, 10),
			// tokens = 10*4 + 0 = 40; floor = 21000 + 40*10 = 21400
			want: 21400,
		},
		{
			name: "mixed: 60 nonzero + 40 zero = 100 bytes",
			data: append(bytes.Repeat([]byte{0x01}, 60), make([]byte, 40)...),
			// tokens = 60*4 + 40 = 280; floor = 21000 + 280*10 = 23800
			want: 23800,
		},
		{
			name: "1000 bytes all nonzero",
			data: bytes.Repeat([]byte{0xab}, 1000),
			// tokens = 1000*4 + 0 = 4000; floor = 21000 + 4000*10 = 61000
			want: 61000,
		},
		{
			name: "1000 bytes all zero",
			data: make([]byte, 1000),
			// tokens = 0*4 + 1000 = 1000; floor = 21000 + 1000*10 = 31000
			want: 31000,
		},
		{
			name: "calldata-heavy tx: 10000 nonzero bytes",
			data: bytes.Repeat([]byte{0xff}, 10000),
			// tokens = 10000*4 = 40000; floor = 21000 + 40000*10 = 421000
			want: 421000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FloorDataGas(tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FloorDataGas() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("FloorDataGas() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestFloorDataGasConsistency verifies the floor data gas is always >= the base TxGas.
func TestFloorDataGasConsistency(t *testing.T) {
	// Any data should produce floor >= TxGas
	testCases := [][]byte{
		nil,
		{},
		{0},
		{1},
		bytes.Repeat([]byte{0x00}, 100),
		bytes.Repeat([]byte{0xff}, 100),
		bytes.Repeat([]byte{0xde, 0xad, 0x00, 0xbe}, 25),
	}
	for _, data := range testCases {
		gas, err := FloorDataGas(data)
		if err != nil {
			t.Fatalf("unexpected error for data len %d: %v", len(data), err)
		}
		if gas < vars.TxGas {
			t.Errorf("FloorDataGas(%d bytes) = %d, below TxGas %d", len(data), gas, vars.TxGas)
		}
	}
}

// TestFloorDataGasVsIntrinsicGas verifies that for data-heavy transactions,
// the floor gas is higher than the intrinsic gas (the whole point of EIP-7623).
func TestFloorDataGasVsIntrinsicGas(t *testing.T) {
	// For large calldata, floor should exceed intrinsic gas.
	// IntrinsicGas with EIP-2028: nz*16 + z*4 + 21000
	// FloorDataGas: (nz*4 + z)*10 + 21000
	// For 1000 nonzero bytes:
	//   Intrinsic = 1000*16 + 21000 = 37000
	//   Floor     = 1000*4*10 + 21000 = 61000  (floor > intrinsic)
	data := bytes.Repeat([]byte{0xff}, 1000)

	intrinsic, err := IntrinsicGas(data, nil, 0, false, true, true, false)
	if err != nil {
		t.Fatal(err)
	}
	floor, err := FloorDataGas(data)
	if err != nil {
		t.Fatal(err)
	}
	if floor <= intrinsic {
		t.Errorf("For 1000 nonzero bytes: floor (%d) should exceed intrinsic (%d)", floor, intrinsic)
	}
}
