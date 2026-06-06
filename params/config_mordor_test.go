package params

import (
	"math/big"
	"testing"
)

// TestOlympiaCoordination_MordorMESS verifies ECIP-1122 Section 3: MESS re-activation
// block must equal the Olympia (EIP-1559) activation block on Mordor testnet.
func TestOlympiaCoordination_MordorMESS(t *testing.T) {
	cfg := MordorChainConfig

	reactivate := cfg.GetECBP1100ReactivateTransition()
	eip1559 := cfg.GetEIP1559Transition()

	if reactivate == nil {
		t.Fatal("ECBP1100ReactivateFBlock is nil — ECIP-1122 requires co-activation with Olympia")
	}
	if eip1559 == nil {
		t.Fatal("EIP1559FBlock is nil")
	}
	if *reactivate != *eip1559 {
		t.Fatalf("ECBP1100ReactivateFBlock (%d) != EIP1559FBlock (%d); ECIP-1122 Section 3 requires co-activation",
			*reactivate, *eip1559)
	}

	// Spiral era (post-deactivation, pre-Olympia): MESS must not be reactivated yet.
	spiralBlock := big.NewInt(11_000_000) // between deactivation (10,400,000) and Olympia
	if cfg.IsEnabled(cfg.GetECBP1100ReactivateTransition, spiralBlock) {
		t.Errorf("MESS should not be reactivated during Spiral era at block %d", spiralBlock)
	}

	// Olympia activation block: MESS must be reactivated.
	olympiaBlock := new(big.Int).SetUint64(*reactivate)
	if !cfg.IsEnabled(cfg.GetECBP1100ReactivateTransition, olympiaBlock) {
		t.Errorf("MESS should be reactivated at Olympia activation block %d", olympiaBlock)
	}
}
