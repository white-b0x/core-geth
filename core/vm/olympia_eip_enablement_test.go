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

package vm

// Olympia EVM opcode enablement tests.
// Verifies the jump table is correctly populated for ETC Olympia EIPs using
// instructionSetForConfig, which is the same table used during live execution.

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/params/types/coregeth"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
)

// newETCVMConfig returns a minimal ETC-style CoreGethChainConfig suitable for VM tests.
// spiralBlock = olympiaBlock/2; all Olympia EIPs activate at olympiaBlock.
func newETCVMConfig(olympiaBlock uint64) *coregeth.CoreGethChainConfig {
	spiralBlock := olympiaBlock / 2
	if spiralBlock == 0 {
		spiralBlock = 1
	}
	return &coregeth.CoreGethChainConfig{
		Ethash:        new(ctypes.EthashConfig),
		EIP3855FBlock: big.NewInt(int64(spiralBlock)),  // PUSH0 (Spiral)
		EIP1559FBlock: big.NewInt(int64(olympiaBlock)), // EIP-1559 (Olympia)
		EIP3198FBlock: big.NewInt(int64(olympiaBlock)), // BASEFEE opcode
		EIP1153FBlock: big.NewInt(int64(olympiaBlock)), // TLOAD / TSTORE
		EIP5656FBlock: big.NewInt(int64(olympiaBlock)), // MCOPY
	}
}

// jumpTableAt returns the jump table for the given config and block number.
func jumpTableAt(cfg *coregeth.CoreGethChainConfig, blockNum int64) *JumpTable {
	return instructionSetForConfig(cfg, false, big.NewInt(blockNum), nil)
}

// opcodeEnabled returns true iff the given opcode maps to a real implementation
// (not opUndefined) in the supplied jump table.
func opcodeEnabled(table *JumpTable, op OpCode) bool {
	return reflect.ValueOf(table[op].execute).Pointer() != reflect.ValueOf(opUndefined).Pointer()
}

const etcOlympiaBlock = 100

// TestBASEFEE_DisabledPreOlympia verifies BASEFEE (0x48) is opUndefined before
// the EIP-3198 activation block.
func TestBASEFEE_DisabledPreOlympia(t *testing.T) {
	cfg := newETCVMConfig(etcOlympiaBlock)
	table := jumpTableAt(cfg, etcOlympiaBlock-1)
	if opcodeEnabled(table, BASEFEE) {
		t.Fatal("BASEFEE must not be available before Olympia activation")
	}
}

// TestBASEFEE_EnabledAtOlympia verifies BASEFEE (0x48) is available at the Olympia
// activation block (EIP-3198).
func TestBASEFEE_EnabledAtOlympia(t *testing.T) {
	cfg := newETCVMConfig(etcOlympiaBlock)
	table := jumpTableAt(cfg, etcOlympiaBlock)
	if !opcodeEnabled(table, BASEFEE) {
		t.Fatal("BASEFEE must be available at Olympia activation (EIP-3198)")
	}
}

// TestTLOAD_DisabledPreOlympia verifies TLOAD (0x5c) is opUndefined before Olympia.
func TestTLOAD_DisabledPreOlympia(t *testing.T) {
	cfg := newETCVMConfig(etcOlympiaBlock)
	table := jumpTableAt(cfg, etcOlympiaBlock-1)
	if opcodeEnabled(table, TLOAD) {
		t.Fatal("TLOAD must not be available before Olympia (EIP-1153)")
	}
}

// TestTLOAD_EnabledAtOlympia verifies TLOAD (0x5c) is available at Olympia (EIP-1153).
func TestTLOAD_EnabledAtOlympia(t *testing.T) {
	cfg := newETCVMConfig(etcOlympiaBlock)
	table := jumpTableAt(cfg, etcOlympiaBlock)
	if !opcodeEnabled(table, TLOAD) {
		t.Fatal("TLOAD must be available at Olympia (EIP-1153)")
	}
}

// TestTSTORE_DisabledPreOlympia verifies TSTORE (0x5d) is opUndefined before Olympia.
func TestTSTORE_DisabledPreOlympia(t *testing.T) {
	cfg := newETCVMConfig(etcOlympiaBlock)
	table := jumpTableAt(cfg, etcOlympiaBlock-1)
	if opcodeEnabled(table, TSTORE) {
		t.Fatal("TSTORE must not be available before Olympia (EIP-1153)")
	}
}

// TestTSTORE_EnabledAtOlympia verifies TSTORE (0x5d) is available at Olympia (EIP-1153).
func TestTSTORE_EnabledAtOlympia(t *testing.T) {
	cfg := newETCVMConfig(etcOlympiaBlock)
	table := jumpTableAt(cfg, etcOlympiaBlock)
	if !opcodeEnabled(table, TSTORE) {
		t.Fatal("TSTORE must be available at Olympia (EIP-1153)")
	}
}

// TestMCOPY_DisabledPreOlympia verifies MCOPY (0x5e) is opUndefined before Olympia.
func TestMCOPY_DisabledPreOlympia(t *testing.T) {
	cfg := newETCVMConfig(etcOlympiaBlock)
	table := jumpTableAt(cfg, etcOlympiaBlock-1)
	if opcodeEnabled(table, MCOPY) {
		t.Fatal("MCOPY must not be available before Olympia (EIP-5656)")
	}
}

// TestMCOPY_EnabledAtOlympia verifies MCOPY (0x5e) is available at Olympia (EIP-5656).
func TestMCOPY_EnabledAtOlympia(t *testing.T) {
	cfg := newETCVMConfig(etcOlympiaBlock)
	table := jumpTableAt(cfg, etcOlympiaBlock)
	if !opcodeEnabled(table, MCOPY) {
		t.Fatal("MCOPY must be available at Olympia (EIP-5656)")
	}
}

// TestBLOBHASH_ExcludedFromETC verifies that BLOBHASH (0x49) is NEVER enabled on ETC.
// ETC does not include EIP-4844 (no blob transactions, no PoS), so BLOBHASH must remain
// opUndefined even after Olympia. If this test fails, the ETC EVM incorrectly
// includes ETH blob logic.
func TestBLOBHASH_ExcludedFromETC(t *testing.T) {
	cfg := newETCVMConfig(etcOlympiaBlock)
	for _, bn := range []int64{0, etcOlympiaBlock - 1, etcOlympiaBlock, etcOlympiaBlock + 1000} {
		table := jumpTableAt(cfg, bn)
		if opcodeEnabled(table, BLOBHASH) {
			t.Fatalf("BLOBHASH must never be enabled on ETC (block %d): ETC does not implement EIP-4844", bn)
		}
	}
}

// TestBLOBBASEFEE_ExcludedFromETC verifies that BLOBBASEFEE (0x4a) is NEVER enabled on
// ETC. Same rationale as TestBLOBHASH_ExcludedFromETC.
func TestBLOBBASEFEE_ExcludedFromETC(t *testing.T) {
	cfg := newETCVMConfig(etcOlympiaBlock)
	for _, bn := range []int64{0, etcOlympiaBlock - 1, etcOlympiaBlock, etcOlympiaBlock + 1000} {
		table := jumpTableAt(cfg, bn)
		if opcodeEnabled(table, BLOBBASEFEE) {
			t.Fatalf("BLOBBASEFEE must never be enabled on ETC (block %d): ETC does not implement EIP-7516", bn)
		}
	}
}

// TestOlympiaAddsAtLeastBASEFEE_TLOAD_TSTORE_MCOPY verifies that all four Olympia-specific
// opcodes are simultaneously available in the Olympia-era jump table.
func TestOlympiaAddsAtLeastBASEFEE_TLOAD_TSTORE_MCOPY(t *testing.T) {
	cfg := newETCVMConfig(etcOlympiaBlock)
	table := jumpTableAt(cfg, etcOlympiaBlock)
	for op, name := range map[OpCode]string{
		BASEFEE: "BASEFEE",
		TLOAD:   "TLOAD",
		TSTORE:  "TSTORE",
		MCOPY:   "MCOPY",
	} {
		if !opcodeEnabled(table, op) {
			t.Errorf("%s (0x%02x) not enabled at Olympia block", name, byte(op))
		}
	}
}

// TestSpiralOpcodes_AreSubsetOfOlympia verifies that any opcode enabled at the Spiral
// block (pre-Olympia) is also enabled at Olympia. Olympia must be a superset of Spiral.
func TestSpiralOpcodes_AreSubsetOfOlympia(t *testing.T) {
	const spiralBlock = etcOlympiaBlock / 2
	cfg := newETCVMConfig(etcOlympiaBlock)
	spiralTable := jumpTableAt(cfg, spiralBlock)
	olympiaTable := jumpTableAt(cfg, etcOlympiaBlock)

	for op := 0; op < 256; op++ {
		if opcodeEnabled(spiralTable, OpCode(op)) && !opcodeEnabled(olympiaTable, OpCode(op)) {
			t.Errorf("opcode 0x%02x is in Spiral but missing from Olympia — Olympia must be a superset", op)
		}
	}
}
