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

package miner

import "testing"

// TestEIP7935DefaultGasCeil verifies the miner default gas ceiling is 60M
// per EIP-7935.
func TestEIP7935DefaultGasCeil(t *testing.T) {
	expected := uint64(60_000_000)
	if DefaultConfig.GasCeil != expected {
		t.Fatalf("miner DefaultConfig.GasCeil = %d, want %d", DefaultConfig.GasCeil, expected)
	}
}
