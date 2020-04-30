package wallet

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/dusk-network/dusk-blockchain/harness/tests"
	"github.com/dusk-network/dusk-protobuf/autogen/go/rusk"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/dusk-network/dusk-blockchain/pkg/core/data/database"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/key"

	"github.com/stretchr/testify/assert"
)

const dbPath = "testDb"

const seedFile = "seed.dat"
const secretFile = "key.dat"

const address = "127.0.0.1:5051"

func TestMain(m *testing.M) {

	//start rusk mock rpc server
	tests.StartMockServer(address)

	// Start all tests
	code := m.Run()

	os.Exit(code)
}

func createRPCConn(t *testing.T) (client rusk.RuskClient, conn *grpc.ClientConn) {
	dialCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var err error
	conn, err = grpc.DialContext(dialCtx, address, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	client = rusk.NewRuskClient(conn)

	return client, conn
}

func TestNewWallet(t *testing.T) {
	netPrefix := byte(1)

	db, err := database.New(dbPath)
	assert.Nil(t, err)
	defer os.RemoveAll(dbPath)
	defer os.Remove(seedFile)
	defer os.Remove(secretFile)

	client, conn := createRPCConn(t)
	defer conn.Close()

	seed, err := GenerateNewSeed(nil)
	require.Nil(t, err)

	ctx := context.Background()
	secretKey, err := client.GenerateSecretKey(ctx, &rusk.GenerateSecretKeyRequest{B: seed})
	require.Nil(t, err)
	require.NotNil(t, secretKey)
	require.NotNil(t, secretKey.A.Data)
	require.NotNil(t, secretKey.B.Data)

	w, err := New(nil, seed, netPrefix, db, "pass", seedFile, secretFile, secretKey)
	assert.Nil(t, err)

	// wrong wallet password
	loadedWallet, err := LoadFromFile(netPrefix, db, "wrongPass", seedFile, secretFile)
	assert.NotNil(t, err)
	assert.Nil(t, loadedWallet)

	// correct wallet password
	loadedWallet, err = LoadFromFile(netPrefix, db, "pass", seedFile, secretFile)
	assert.Nil(t, err)

	assert.Equal(t, w.SecretKey().A.Data, loadedWallet.SecretKey().A.Data)
	assert.Equal(t, w.SecretKey().B.Data, loadedWallet.SecretKey().B.Data)

	assert.Equal(t, w.consensusKeys.BLSSecretKey, loadedWallet.consensusKeys.BLSSecretKey)
	assert.True(t, bytes.Equal(w.consensusKeys.BLSPubKeyBytes, loadedWallet.consensusKeys.BLSPubKeyBytes))
}

func TestCatchEOF(t *testing.T) {
	netPrefix := byte(1)

	client, conn := createRPCConn(t)
	defer conn.Close()

	db, err := database.New(dbPath)
	assert.Nil(t, err)
	defer os.RemoveAll(dbPath)

	defer os.Remove(seedFile)
	defer os.Remove(secretFile)

	// Generate 1000 new wallets
	for i := 0; i < 1000; i++ {
		seed, err := GenerateNewSeed(nil)
		require.Nil(t, err)

		ctx := context.Background()
		secretKey, err := client.GenerateSecretKey(ctx, &rusk.GenerateSecretKeyRequest{B: seed})
		require.Nil(t, err)

		_, err = New(nil, seed, netPrefix, db, "pass", seedFile, secretFile, secretKey)
		assert.Nil(t, err)
		os.Remove(seedFile)
		os.Remove(secretFile)
	}
}

func generateWallet(t *testing.T, netPrefix byte, walletPath, seedFile, secretFile string) *Wallet { //nolint:unparam
	db, err := database.New(walletPath)
	assert.Nil(t, err)
	//defer os.RemoveAll(walletPath)

	client, conn := createRPCConn(t)
	defer conn.Close()

	seed, err := GenerateNewSeed(nil)
	require.Nil(t, err)

	ctx := context.Background()
	secretKey, err := client.GenerateSecretKey(ctx, &rusk.GenerateSecretKeyRequest{B: seed})
	require.Nil(t, err)

	w, err := New(nil, seed, netPrefix, db, "pass", seedFile, secretFile, secretKey)
	assert.Nil(t, err)
	return w
}

func generateSendAddr(t *testing.T, netPrefix byte, randKeyPair *key.Key) key.PublicAddress {
	pubAddr, err := randKeyPair.PublicKey().PublicAddress(netPrefix)
	assert.Nil(t, err)
	return *pubAddr
}
