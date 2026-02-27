//go:build live

package live

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
)

// getFukuiiRPC returns the Fukuii RPC URL from env or default.
func getFukuiiRPC() string {
	if url := os.Getenv("FUKUII_RPC"); url != "" {
		return url
	}
	return "http://localhost:8553"
}

// getBesuRPC returns the Besu RPC URL from env or default.
func getBesuRPC() string {
	if url := os.Getenv("BESU_RPC"); url != "" {
		return url
	}
	return "http://localhost:8548"
}

// tryDialRPC attempts to connect, returning nil if it fails.
func tryDialRPC(url string) *rpc.Client {
	client, err := rpc.Dial(url)
	if err != nil {
		return nil
	}
	return client
}

// TestCrossClientForkBlockHash compares the fork block hash across available clients.
func TestCrossClientForkBlockHash(t *testing.T) {
	mordorClient := dialRPC(t, getMordorRPC())
	defer mordorClient.Close()

	requireForkReached(t, mordorClient, MordorForkBlock)

	coreGethBlock := getBlockByNumber(t, mordorClient, MordorForkBlock, false)
	t.Logf("core-geth fork block hash: %s", coreGethBlock.Hash.Hex())

	// Try Fukuii
	if fukuiiClient := tryDialRPC(getFukuiiRPC()); fukuiiClient != nil {
		defer fukuiiClient.Close()
		fukuiiBlock := getBlockByNumber(t, fukuiiClient, MordorForkBlock, false)
		if fukuiiBlock.Hash != coreGethBlock.Hash {
			t.Errorf("CONSENSUS FAILURE: core-geth fork block hash %s != fukuii %s",
				coreGethBlock.Hash.Hex(), fukuiiBlock.Hash.Hex())
		} else {
			t.Logf("fukuii fork block hash matches: %s", fukuiiBlock.Hash.Hex())
		}
	} else {
		t.Log("fukuii not available — skipping cross-client check")
	}

	// Try Besu
	if besuClient := tryDialRPC(getBesuRPC()); besuClient != nil {
		defer besuClient.Close()
		besuBlock := getBlockByNumber(t, besuClient, MordorForkBlock, false)
		if besuBlock.Hash != coreGethBlock.Hash {
			t.Errorf("CONSENSUS FAILURE: core-geth fork block hash %s != besu %s",
				coreGethBlock.Hash.Hex(), besuBlock.Hash.Hex())
		} else {
			t.Logf("besu fork block hash matches: %s", besuBlock.Hash.Hex())
		}
	} else {
		t.Log("besu not available — skipping cross-client check")
	}
}

// TestCrossClientStateRoot compares state roots at fork+10 across available clients.
func TestCrossClientStateRoot(t *testing.T) {
	mordorClient := dialRPC(t, getMordorRPC())
	defer mordorClient.Close()

	requireForkReached(t, mordorClient, MordorForkBlock+10)

	checkBlock := uint64(MordorForkBlock + 10)
	coreGethBlock := getBlockByNumber(t, mordorClient, checkBlock, false)
	t.Logf("core-geth block %d stateRoot: %s", checkBlock, coreGethBlock.StateRoot.Hex())

	// Try Fukuii
	if fukuiiClient := tryDialRPC(getFukuiiRPC()); fukuiiClient != nil {
		defer fukuiiClient.Close()
		fukuiiBlock := getBlockByNumber(t, fukuiiClient, checkBlock, false)
		if fukuiiBlock.StateRoot != coreGethBlock.StateRoot {
			t.Errorf("STATE ROOT MISMATCH at block %d: core-geth %s != fukuii %s",
				checkBlock, coreGethBlock.StateRoot.Hex(), fukuiiBlock.StateRoot.Hex())
		} else {
			t.Logf("fukuii state root matches at block %d", checkBlock)
		}
	} else {
		t.Log("fukuii not available — skipping state root check")
	}

	// Try Besu
	if besuClient := tryDialRPC(getBesuRPC()); besuClient != nil {
		defer besuClient.Close()
		besuBlock := getBlockByNumber(t, besuClient, checkBlock, false)
		if besuBlock.StateRoot != coreGethBlock.StateRoot {
			t.Errorf("STATE ROOT MISMATCH at block %d: core-geth %s != besu %s",
				checkBlock, coreGethBlock.StateRoot.Hex(), besuBlock.StateRoot.Hex())
		} else {
			t.Logf("besu state root matches at block %d", checkBlock)
		}
	} else {
		t.Log("besu not available — skipping state root check")
	}
}
