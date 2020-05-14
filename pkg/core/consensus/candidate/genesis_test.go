package candidate

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/dusk-network/dusk-blockchain/pkg/config"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/block"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/transactions"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/message"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/topics"
	"github.com/dusk-network/dusk-blockchain/pkg/util/nativeutils/rpcbus"
	"github.com/stretchr/testify/require"
)

func TestGenerateGenesis(t *testing.T) {
	// FIXME: 417 - KEYS the key generation from the SEED uses rusk atm. Also, the
	// genesis needs to be regenerated accordingly. This test should be
	// adjusted to become an integration test
	rpcBus := rpcbus.New()

	go provideMempoolTxs(rpcBus)

	seed := make([]byte, 64)
	if _, err := rand.Read(seed); err != nil {
		t.Fatal(err.Error())
	}

	pk := "61c36e407ac91f20174572eec95f692f5cff1c40bacd1b9f86c7fa7202e93bb6753c2f424caf3c9220876e8cfe0afdff7ffd7c984d5c7d95fa0b46cf3781d883"
	pkBytes, err := hex.DecodeString(pk)
	require.Nil(t, err)

	publicKey := &transactions.PublicKey{
		AG: &transactions.CompressedPoint{
			Y: pkBytes[0:32],
		},
		BG: &transactions.CompressedPoint{
			Y: pkBytes[32:64],
		},
	}

	// Generate a new genesis block with new wallet pubkey
	genesisHex, err := GenerateGenesisBlock(rpcBus, publicKey)
	if err != nil {
		fmt.Println(err)
		t.Fatalf("expecting valid genesis block")
	}

	// Decode the result hex value to ensure it's a valid block
	blob, err := hex.DecodeString(genesisHex)
	if err != nil {
		t.Fatalf("expecting valid hex %s", err.Error())
	}

	var buf bytes.Buffer
	buf.Write(blob)

	b := block.NewBlock()
	if err := message.UnmarshalBlock(&buf, b); err != nil {
		t.Fatalf("expecting decodable hex %s", err.Error())
	}

	// Print blob
	t.Logf("genesis: %s", genesisHex)
}

func TestGenesisBlock(t *testing.T) {
	// read the hard-coded genesis blob for testnet
	blob, err := hex.DecodeString(config.TestNetGenesisBlob)
	if err != nil {
		t.Fatalf("expecting valid cfg.TestNetGenesisBlob %s", err.Error())
	}

	// decode the blob to a block
	var buf bytes.Buffer
	buf.Write(blob)

	b := block.NewBlock()
	if err := message.UnmarshalBlock(&buf, b); err != nil {
		t.Fatalf("expecting decodable cfg.TestNetGenesisBlob %s", err.Error())
	}

	// sanity checks
	if b.Header.Height != 0 {
		t.Fatalf("expecting valid height in cfg.TestNetGenesisBlob")
	}

	if b.Header.Version != 0 {
		t.Fatalf("expecting valid version in cfg.TestNetGenesisBlob")
	}

	if b.Txs[0].Type() != transactions.Distribute {
		t.Fatalf("expecting coinbase tx in cfg.TestNetGenesisBlob")
	}
}

func provideMempoolTxs(rpcBus *rpcbus.RPCBus) {
	c := make(chan rpcbus.Request, 1)
	if err := rpcBus.Register(topics.GetMempoolTxsBySize, c); err != nil {
		panic(err)
	}

	r := <-c
	r.RespChan <- rpcbus.NewResponse([]transactions.ContractCall{}, nil)
}
