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

// Olympia EVM opcode execution tests.
// These tests go beyond jump-table membership (covered in olympia_eip_enablement_test.go)
// and verify correct runtime behaviour: BASEFEE pushes the right value, costs 2 gas, etc.
// TLOAD/TSTORE and MCOPY execution is also verified.

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

// newOpcodeTestEVM creates a minimal EVM + interpreter for opcode execution tests.
// cfg selects the jump table; blockNum and baseFee are placed in BlockContext.
func newOpcodeTestEVM(t *testing.T, blockNum int64, baseFee *big.Int) (*EVM, *EVMInterpreter) {
	t.Helper()
	statedb, err := state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	if err != nil {
		t.Fatalf("state.New: %v", err)
	}
	cfg := newETCVMConfig(etcOlympiaBlock)
	ctx := BlockContext{
		BlockNumber: big.NewInt(blockNum),
		BaseFee:     baseFee,
		Transfer:    func(StateDB, common.Address, common.Address, *uint256.Int) {},
	}
	evm := NewEVM(ctx, TxContext{GasPrice: big.NewInt(0)}, statedb, cfg, Config{})
	interp := NewEVMInterpreter(evm)
	evm.interpreter = interp
	return evm, interp
}

// newOpcodeScope creates a minimal ScopeContext with a fresh stack, memory, and contract.
func newOpcodeScope(gas uint64) ScopeContext {
	contract := NewContract(AccountRef(common.Address{}), AccountRef(common.Address{1}), new(uint256.Int), gas)
	return ScopeContext{
		Memory:   NewMemory(),
		Stack:    newstack(),
		Contract: contract,
	}
}

// TestBASEFEE_CostsGasQuickStep verifies that BASEFEE (0x48) has constantGas == GasQuickStep (2).
func TestBASEFEE_CostsGasQuickStep(t *testing.T) {
	table := jumpTableAt(newETCVMConfig(etcOlympiaBlock), etcOlympiaBlock)
	if table[BASEFEE] == nil {
		t.Fatal("BASEFEE not in jump table at Olympia")
	}
	if table[BASEFEE].constantGas != GasQuickStep {
		t.Fatalf("BASEFEE constantGas = %d, want GasQuickStep = %d", table[BASEFEE].constantGas, GasQuickStep)
	}
}

// TestBASEFEE_PushesCorrectValue verifies that BASEFEE pushes the block's baseFee onto the stack.
func TestBASEFEE_PushesCorrectValue(t *testing.T) {
	const wantBaseFee = uint64(42)
	_, interp := newOpcodeTestEVM(t, etcOlympiaBlock, new(big.Int).SetUint64(wantBaseFee))
	scope := newOpcodeScope(100)
	pc := uint64(0)
	if _, err := opBaseFee(&pc, interp, &scope); err != nil {
		t.Fatalf("opBaseFee: %v", err)
	}
	if scope.Stack.len() != 1 {
		t.Fatalf("stack len = %d, want 1", scope.Stack.len())
	}
	got := scope.Stack.peek().Uint64()
	if got != wantBaseFee {
		t.Fatalf("BASEFEE pushed %d, want %d", got, wantBaseFee)
	}
}

// TestBASEFEE_ZeroWhenBaseFeeZero verifies BASEFEE pushes 0 when the block baseFee is 0.
func TestBASEFEE_ZeroWhenBaseFeeZero(t *testing.T) {
	_, interp := newOpcodeTestEVM(t, etcOlympiaBlock, common.Big0)
	scope := newOpcodeScope(100)
	pc := uint64(0)
	if _, err := opBaseFee(&pc, interp, &scope); err != nil {
		t.Fatalf("opBaseFee: %v", err)
	}
	got := scope.Stack.peek().Uint64()
	if got != 0 {
		t.Fatalf("BASEFEE with zero baseFee pushed %d, want 0", got)
	}
}

// TestBASEFEE_HandlesLargeValue verifies BASEFEE handles a 1 Gwei baseFee without error.
func TestBASEFEE_HandlesLargeValue(t *testing.T) {
	const oneGwei = uint64(1_000_000_000)
	_, interp := newOpcodeTestEVM(t, etcOlympiaBlock, new(big.Int).SetUint64(oneGwei))
	scope := newOpcodeScope(100)
	pc := uint64(0)
	if _, err := opBaseFee(&pc, interp, &scope); err != nil {
		t.Fatalf("opBaseFee(1 Gwei): %v", err)
	}
	got := scope.Stack.peek().Uint64()
	if got != oneGwei {
		t.Fatalf("BASEFEE pushed %d, want %d", got, oneGwei)
	}
}

// TestBASEFEE_UnavailablePreOlympia verifies BASEFEE (0x48) is opUndefined before Olympia.
// Complements TestBASEFEE_DisabledPreOlympia in olympia_eip_enablement_test.go with
// a direct table[BASEFEE].constantGas check that distinguishes opUndefined from a real op.
func TestBASEFEE_UnavailablePreOlympia(t *testing.T) {
	table := jumpTableAt(newETCVMConfig(etcOlympiaBlock), etcOlympiaBlock-1)
	if opcodeEnabled(table, BASEFEE) {
		t.Fatal("BASEFEE must not be available before Olympia")
	}
	// opUndefined has constantGas == 0
	if table[BASEFEE] != nil && table[BASEFEE].constantGas != 0 {
		t.Fatalf("pre-Olympia BASEFEE constantGas = %d, want 0 (opUndefined)", table[BASEFEE].constantGas)
	}
}

// TestTLOAD_TSTORE_RoundTrip verifies that TSTORE followed by TLOAD retrieves the stored value.
func TestTLOAD_TSTORE_RoundTrip(t *testing.T) {
	evm, interp := newOpcodeTestEVM(t, etcOlympiaBlock, common.Big0)
	caller := common.Address{}
	to := common.Address{1}
	evm.StateDB.CreateAccount(caller)
	evm.StateDB.CreateAccount(to)
	contract := NewContract(AccountRef(caller), AccountRef(to), new(uint256.Int), 0)
	scope := ScopeContext{Memory: NewMemory(), Stack: newstack(), Contract: contract}
	pc := uint64(0)

	value := new(uint256.Int).SetUint64(0xdeadbeef)
	slot := new(uint256.Int).SetUint64(0)

	// TSTORE: push value then slot (slot is consumed first by TSTORE)
	scope.Stack.push(value)
	scope.Stack.push(slot)
	if _, err := opTstore(&pc, interp, &scope); err != nil {
		t.Fatalf("opTstore: %v", err)
	}
	if scope.Stack.len() != 0 {
		t.Fatalf("stack after TSTORE: len=%d, want 0", scope.Stack.len())
	}

	// TLOAD: push slot, expect value on stack
	scope.Stack.push(new(uint256.Int).SetUint64(0))
	if _, err := opTload(&pc, interp, &scope); err != nil {
		t.Fatalf("opTload: %v", err)
	}
	if scope.Stack.len() != 1 {
		t.Fatalf("stack after TLOAD: len=%d, want 1", scope.Stack.len())
	}
	got := scope.Stack.peek()
	if got.Cmp(value) != 0 {
		t.Fatalf("TLOAD got %s, want %s", got, value)
	}
}

// TestTLOAD_TSTORE_UnavailablePreOlympia verifies TLOAD (0x5c) and TSTORE (0x5d) are
// opUndefined before Olympia. Combines the two checks since both activate on EIP-1153.
func TestTLOAD_TSTORE_UnavailablePreOlympia(t *testing.T) {
	table := jumpTableAt(newETCVMConfig(etcOlympiaBlock), etcOlympiaBlock-1)
	for op, name := range map[OpCode]string{TLOAD: "TLOAD", TSTORE: "TSTORE"} {
		if opcodeEnabled(table, op) {
			t.Errorf("%s (0x%02x) must not be available before Olympia", name, byte(op))
		}
	}
}

// TestMCOPY_CopiesMemory verifies that MCOPY correctly copies a memory region.
func TestMCOPY_CopiesMemory(t *testing.T) {
	_, interp := newOpcodeTestEVM(t, etcOlympiaBlock, common.Big0)
	scope := newOpcodeScope(10000)
	scope.Memory.Resize(128)
	// Write source data: 32 bytes of 0xAB at offset 32
	for i := 32; i < 64; i++ {
		scope.Memory.store[i] = 0xAB
	}
	pc := uint64(0)
	// MCOPY(dst=0, src=32, len=32): stack order is len, src, dst (top is dst)
	scope.Stack.push(new(uint256.Int).SetUint64(32)) // len
	scope.Stack.push(new(uint256.Int).SetUint64(32)) // src
	scope.Stack.push(new(uint256.Int).SetUint64(0))  // dst
	if _, err := opMcopy(&pc, interp, &scope); err != nil {
		t.Fatalf("opMcopy: %v", err)
	}
	// Verify bytes 0-31 now contain 0xAB
	for i := 0; i < 32; i++ {
		if scope.Memory.store[i] != 0xAB {
			t.Fatalf("MCOPY: memory[%d] = 0x%02x, want 0xAB", i, scope.Memory.store[i])
		}
	}
}

// TestMCOPY_UnavailablePreOlympia verifies MCOPY (0x5e) is opUndefined before Olympia.
func TestMCOPY_UnavailablePreOlympia(t *testing.T) {
	table := jumpTableAt(newETCVMConfig(etcOlympiaBlock), etcOlympiaBlock-1)
	if opcodeEnabled(table, MCOPY) {
		t.Fatal("MCOPY must not be available before Olympia")
	}
}
