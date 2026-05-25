// Copyright 2024 The go-ethereum Authors
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

package types

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
)

// DelegationPrefix is used to identify delegation designations in account code.
// An account with code starting with this prefix is a delegated account.
var DelegationPrefix = []byte{0xef, 0x01, 0x00}

// ParseDelegation tries to parse the 23-byte delegation code of an account and
// return the target address if valid.
func ParseDelegation(b []byte) (common.Address, bool) {
	if len(b) != 23 || !bytes.HasPrefix(b, DelegationPrefix) {
		return common.Address{}, false
	}
	var addr common.Address
	copy(addr[:], b[3:])
	return addr, true
}

// AddressToDelegation wraps an address with the delegation prefix.
func AddressToDelegation(addr common.Address) []byte {
	return append(DelegationPrefix, addr.Bytes()...)
}

// SetCodeTx implements the EIP-7702 transaction type that temporarily
// sets code on an EOA.
type SetCodeTx struct {
	ChainID    *uint256.Int
	Nonce      uint64
	GasTipCap  *uint256.Int // a.k.a. maxPriorityFeePerGas
	GasFeeCap  *uint256.Int // a.k.a. maxFeePerGas
	Gas        uint64
	To         common.Address
	Value      *uint256.Int
	Data       []byte
	AccessList AccessList
	AuthList   []SetCodeAuthorization

	// Signature values
	V *uint256.Int `json:"v" gencodec:"required"`
	R *uint256.Int `json:"r" gencodec:"required"`
	S *uint256.Int `json:"s" gencodec:"required"`
}

// SetCodeAuthorization is an authorization from an account to set its code.
type SetCodeAuthorization struct {
	ChainID uint256.Int    `json:"chainId" gencodec:"required"`
	Address common.Address `json:"address" gencodec:"required"`
	Nonce   uint64         `json:"nonce"   gencodec:"required"`
	V       uint8          `json:"yParity" gencodec:"required"`
	R       uint256.Int    `json:"r"       gencodec:"required"`
	S       uint256.Int    `json:"s"       gencodec:"required"`
}

// Authority recovers the authorizing account address from the signature.
func (a *SetCodeAuthorization) Authority() (common.Address, error) {
	sighash := a.sigHash()
	v := a.V
	if v != 0 && v != 1 {
		return common.Address{}, errors.New("invalid yParity value")
	}
	// Construct signature: R (32 bytes) + S (32 bytes) + V (1 byte)
	var sig [65]byte
	r := a.R.Bytes32()
	s := a.S.Bytes32()
	copy(sig[0:32], r[:])
	copy(sig[32:64], s[:])
	sig[64] = v

	// Recover public key
	pub, err := crypto.Ecrecover(sighash[:], sig[:])
	if err != nil {
		return common.Address{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return common.Address{}, errors.New("invalid public key")
	}
	var addr common.Address
	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	return addr, nil
}

// sigHash returns the hash to be signed by the authority.
func (a *SetCodeAuthorization) sigHash() common.Hash {
	return prefixedRlpHash(0x05, []any{
		a.ChainID,
		a.Address,
		a.Nonce,
	})
}

// SignSetCode signs an authorization with the given private key.
func SignSetCode(prv *ecdsa.PrivateKey, auth SetCodeAuthorization) (SetCodeAuthorization, error) {
	h := auth.sigHash()
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return auth, err
	}
	auth.R = *uint256.NewInt(0).SetBytes(sig[:32])
	auth.S = *uint256.NewInt(0).SetBytes(sig[32:64])
	auth.V = sig[64]
	return auth, nil
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *SetCodeTx) copy() TxData {
	cpy := &SetCodeTx{
		Nonce: tx.Nonce,
		To:    tx.To,
		Data:  common.CopyBytes(tx.Data),
		Gas:   tx.Gas,
		// These are copied below.
		AccessList: make(AccessList, len(tx.AccessList)),
		AuthList:   make([]SetCodeAuthorization, len(tx.AuthList)),
		Value:      new(uint256.Int),
		ChainID:    new(uint256.Int),
		GasTipCap:  new(uint256.Int),
		GasFeeCap:  new(uint256.Int),
		V:          new(uint256.Int),
		R:          new(uint256.Int),
		S:          new(uint256.Int),
	}
	copy(cpy.AccessList, tx.AccessList)
	copy(cpy.AuthList, tx.AuthList)

	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.GasTipCap != nil {
		cpy.GasTipCap.Set(tx.GasTipCap)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	return cpy
}

// accessors for innerTx.
func (tx *SetCodeTx) txType() byte              { return SetCodeTxType }
func (tx *SetCodeTx) chainID() *big.Int         { return tx.ChainID.ToBig() }
func (tx *SetCodeTx) accessList() AccessList    { return tx.AccessList }
func (tx *SetCodeTx) data() []byte              { return tx.Data }
func (tx *SetCodeTx) gas() uint64               { return tx.Gas }
func (tx *SetCodeTx) gasFeeCap() *big.Int       { return tx.GasFeeCap.ToBig() }
func (tx *SetCodeTx) gasTipCap() *big.Int       { return tx.GasTipCap.ToBig() }
func (tx *SetCodeTx) gasPrice() *big.Int        { return tx.GasFeeCap.ToBig() }
func (tx *SetCodeTx) value() *big.Int           { return tx.Value.ToBig() }
func (tx *SetCodeTx) nonce() uint64             { return tx.Nonce }
func (tx *SetCodeTx) to() *common.Address       { tmp := tx.To; return &tmp }
func (tx *SetCodeTx) blobGas() uint64           { return 0 }
func (tx *SetCodeTx) blobGasFeeCap() *big.Int   { return nil }
func (tx *SetCodeTx) blobHashes() []common.Hash { return nil }

func (tx *SetCodeTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap.ToBig())
	}
	tip := dst.Sub(tx.GasFeeCap.ToBig(), baseFee)
	if tip.Cmp(tx.GasTipCap.ToBig()) > 0 {
		tip.Set(tx.GasTipCap.ToBig())
	}
	return tip.Add(tip, baseFee)
}

func (tx *SetCodeTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V.ToBig(), tx.R.ToBig(), tx.S.ToBig()
}

func (tx *SetCodeTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID.SetFromBig(chainID)
	tx.V.SetFromBig(v)
	tx.R.SetFromBig(r)
	tx.S.SetFromBig(s)
}

func (tx *SetCodeTx) encode(b *bytes.Buffer) error {
	return rlp.Encode(b, tx)
}

func (tx *SetCodeTx) decode(input []byte) error {
	return rlp.DecodeBytes(input, tx)
}
