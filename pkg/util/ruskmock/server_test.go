// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT License was not distributed with this
// file, you can obtain one at https://opensource.org/licenses/MIT.
//
// Copyright (c) DUSK NETWORK. All rights reserved.

package ruskmock

import (
	"bytes"
	"context"
	"os"
	"testing"

	ristretto "github.com/bwesterb/go-ristretto"
	"github.com/dusk-network/dusk-blockchain/pkg/config"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/user"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/ipc/transactions"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/encoding"
	"github.com/dusk-network/dusk-blockchain/pkg/rpc/client"
	"github.com/dusk-network/dusk-protobuf/autogen/go/rusk"
	zkproof "github.com/dusk-network/dusk-zkproof"
	"github.com/stretchr/testify/assert"
)

const walletDBName = "walletDB"

func TestGetProvisioners(t *testing.T) {
	s := setupRuskMockTest(t, DefaultConfig())
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateStateClient(ctx, "localhost:10000")

	resp, err := c.GetProvisioners(ctx, &rusk.GetProvisionersRequest{})
	assert.NoError(t, err)

	p := user.NewProvisioners()
	memberMap := make(map[string]*user.Member)

	for i := range resp.Provisioners {
		member := new(user.Member)
		transactions.UMember(resp.Provisioners[i], member)

		memberMap[string(member.PublicKeyBLS)] = member

		p.Set.Insert(member.PublicKeyBLS)
	}

	p.Members = memberMap

	// Should have gotten an identical set
	assert.Equal(t, s.p, p)
}

func TestVerifyStateTransition(t *testing.T) {
	s := setupRuskMockTest(t, DefaultConfig())
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateStateClient(ctx, "localhost:10000")

	resp, err := c.VerifyStateTransition(ctx, &rusk.VerifyStateTransitionRequest{})
	assert.NoError(t, err)

	// Should have gotten an empty FailedCalls slice
	assert.Empty(t, resp.FailedCalls)
}

func TestFailedVerifyStateTransition(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PassStateTransitionValidation = false

	s := setupRuskMockTest(t, cfg)
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateStateClient(ctx, "localhost:10000")

	// Send a request with 5 calls
	calls := make([]*rusk.Transaction, 5)
	for i := range calls {
		call := new(rusk.Transaction)
		tx := transactions.RandTx()

		transactions.MTransaction(call, tx)

		calls[i] = call
	}

	resp, err := c.VerifyStateTransition(ctx, &rusk.VerifyStateTransitionRequest{Txs: calls})
	assert.NoError(t, err)

	// FailedCalls should be a slice of numbers 0 to 4
	assert.Equal(t, 5, len(resp.FailedCalls))

	for i, n := range resp.FailedCalls {
		assert.Equal(t, uint64(i), n)
	}
}

func TestExecuteStateTransition(t *testing.T) {
	s := setupRuskMockTest(t, DefaultConfig())
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateStateClient(ctx, "localhost:10000")

	// We execute the state transition with a single stake transaction, to ensure
	// the provisioners are updated.
	sc, _ := client.CreateStakeClient(ctx, "localhost:10000")
	value := uint64(712389)

	tx, err := sc.NewStake(ctx, &rusk.StakeTransactionRequest{
		Value:        value,
		PublicKeyBls: s.w.ConsensusKeys().BLSPubKeyBytes,
	})
	assert.NoError(t, err)

	resp, err := c.ExecuteStateTransition(ctx, &rusk.ExecuteStateTransitionRequest{
		Txs:    []*rusk.Transaction{tx},
		Height: 1,
	})
	assert.NoError(t, err)

	assert.True(t, resp.Success)

	// Check that the provisioner set has a stake for our public key, with the
	// chosen value
	m := s.p.Members[string(s.w.ConsensusKeys().BLSPubKeyBytes)]
	assert.Equal(t, value, m.Stakes[0].Amount)
}

func TestFailedExecuteStateTransition(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PassStateTransition = false

	s := setupRuskMockTest(t, cfg)
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateStateClient(ctx, "localhost:10000")

	resp, err := c.ExecuteStateTransition(ctx, &rusk.ExecuteStateTransitionRequest{})
	assert.NoError(t, err)

	assert.False(t, resp.Success)
}

func TestGenerateScore(t *testing.T) {
	s := setupRuskMockTest(t, DefaultConfig())
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateBlindBidServiceClient(ctx, "localhost:10000")

	resp, err := c.GenerateScore(ctx, &rusk.GenerateScoreRequest{})
	assert.NoError(t, err)

	// Ensure the returned score has all fields populated
	assert.NotEmpty(t, resp.BlindbidProof)
	assert.NotEmpty(t, resp.Score)
	assert.NotEmpty(t, resp.ProverIdentity)
}

func TestVerifyScore(t *testing.T) {
	s := setupRuskMockTest(t, DefaultConfig())
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateBlindBidServiceClient(ctx, "localhost:10000")

	resp, err := c.VerifyScore(ctx, &rusk.VerifyScoreRequest{})
	assert.NoError(t, err)

	// Should have gotten `true`
	assert.True(t, resp.Success)
}

func TestFailedVerifyScore(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PassScoreValidation = false

	s := setupRuskMockTest(t, cfg)
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateBlindBidServiceClient(ctx, "localhost:10000")

	resp, err := c.VerifyScore(ctx, &rusk.VerifyScoreRequest{})
	assert.NoError(t, err)

	// Should have gotten `false`
	assert.False(t, resp.Success)
}

func TestGenerateStealthAddress(t *testing.T) {
	s := setupRuskMockTest(t, DefaultConfig())
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateKeysClient(ctx, "localhost:10000")

	// The server will just generate a stealth address for it's own public
	// key, so we don't need to send a proper one.
	resp, err := c.GenerateStealthAddress(ctx, &rusk.PublicKey{})
	assert.NoError(t, err)

	// Since we don't know the randomness with which the stealth address was
	// created, we can't really do an equality check. Let's at least make sure
	// that it is a valid ristretto point.
	var pBytes [32]byte
	var p ristretto.Point

	copy(pBytes[:], resp.RG[:])
	assert.True(t, p.SetBytes(&pBytes))
}

func TestGenerateKeys(t *testing.T) {
	s := setupRuskMockTest(t, DefaultConfig())
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateKeysClient(ctx, "localhost:10000")

	resp, err := c.GenerateKeys(ctx, &rusk.GenerateKeysRequest{})
	assert.NoError(t, err)

	var spendBytes [32]byte
	var pSpend ristretto.Scalar

	copy(spendBytes[:], resp.Sk.A[:])
	pSpend.SetBytes(&spendBytes)

	sPSpendBytes, err := s.w.PrivateSpend()
	assert.NoError(t, err)

	var sPSpend ristretto.Scalar

	assert.NoError(t, sPSpend.UnmarshalBinary(sPSpendBytes))
	assert.True(t, sPSpend.Equals(&pSpend))

	// Since we don't know the randomness with which the stealth address was
	// created, we can't really do an equality check. Let's at least make sure
	// that it is a valid ristretto point.
	var pBytes [32]byte
	var p ristretto.Point

	copy(pBytes[:], resp.Pk.AG[:])
	assert.True(t, p.SetBytes(&pBytes))
}

// TODO: check values for correctness
// This is currently quite hard to do, so I will defer it for now.
// In any case, if the call succeeds, we know we've successfully
// created a transaction and marshaled it, so checking values
// would be icing on top of the cake.
func TestNewTransfer(t *testing.T) {
	s := setupRuskMockTest(t, DefaultConfig())
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateTransferClient(ctx, "localhost:10000")

	// Send a transfer to ourselves
	var r ristretto.Scalar
	r.Rand()

	pk := s.w.PublicKey()

	_, err := c.NewTransfer(ctx, &rusk.TransferTransactionRequest{
		Value:     100,
		Recipient: append(pk.PubSpend.Bytes(), pk.PubView.Bytes()...),
	})
	assert.NoError(t, err)
}

func TestNewStake(t *testing.T) {
	s := setupRuskMockTest(t, DefaultConfig())
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateStakeClient(ctx, "localhost:10000")

	resp, err := c.NewStake(ctx, &rusk.StakeTransactionRequest{
		Value:        100,
		PublicKeyBls: s.w.ConsensusKeys().BLSPubKeyBytes,
	})
	assert.NoError(t, err)

	// The bls public key and locktime should be included in the calldata.
	var locktime uint64

	plBuf := bytes.NewBuffer(resp.Payload)

	pl := transactions.NewTransactionPayload()
	assert.NoError(t, transactions.UnmarshalTransactionPayload(plBuf, pl))

	buf := bytes.NewBuffer(pl.CallData)

	err = encoding.ReadUint64LE(buf, &locktime)
	assert.NoError(t, err)

	assert.Equal(t, uint64(250000), locktime)

	pkBLS := make([]byte, 0)

	err = encoding.ReadVarBytes(buf, &pkBLS)
	assert.NoError(t, err)

	assert.Equal(t, s.w.ConsensusKeys().BLSPubKeyBytes, pkBLS)
	assert.Equal(t, uint32(4), resp.Type)
}

func TestNewBid(t *testing.T) {
	s := setupRuskMockTest(t, DefaultConfig())
	defer cleanup(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, _ := client.CreateBidServiceClient(ctx, "localhost:10000")

	// Generate a K
	var k ristretto.Scalar
	k.Rand()

	resp, err := c.NewBid(ctx, &rusk.BidTransactionRequest{
		K:     k.Bytes(),
		Value: 100,
	})
	assert.NoError(t, err)

	// The M and locktime should be encoded in the call data
	var locktime uint64

	plBuf := bytes.NewBuffer(resp.Tx.Payload)

	pl := transactions.NewTransactionPayload()
	assert.NoError(t, transactions.UnmarshalTransactionPayload(plBuf, pl))

	buf := bytes.NewBuffer(pl.CallData)

	err = encoding.ReadUint64LE(buf, &locktime)
	assert.NoError(t, err)

	assert.Equal(t, uint64(250000), locktime)

	m := zkproof.CalculateM(k)
	mBytes := make([]byte, 32)

	err = encoding.Read256(buf, mBytes)
	assert.NoError(t, err)

	assert.Equal(t, m.Bytes(), mBytes)
	assert.Equal(t, uint32(3), resp.Tx.Type)
}

func setupRuskMockTest(t *testing.T, cfg *Config) *Server {
	c := config.Registry{}
	// Hardcode wallet values, so that it always starts up correctly
	c.Wallet.Store = walletDBName
	c.Wallet.File = "../../../devnet-wallets/wallet0.dat"

	s, err := New(cfg, c)
	assert.NoError(t, err)

	assert.NoError(t, s.Serve("tcp", ":10000"))
	return s
}

func cleanup(s *Server) {
	_ = s.Stop()

	if err := os.RemoveAll(walletDBName + "_2"); err != nil {
		panic(err)
	}
}
