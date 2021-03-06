// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT License was not distributed with this
// file, you can obtain one at https://opensource.org/licenses/MIT.
//
// Copyright (c) DUSK NETWORK. All rights reserved.

package rpc_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/dusk-network/dusk-blockchain/pkg/rpc"
	"github.com/stretchr/testify/require"
)

func TestSignAuthToken(t *testing.T) {
	assert := require.New(t)
	pk, sk, _ := ed25519.GenerateKey(rand.Reader)

	tky := rpc.AuthToken{
		AccessToken: "access",
		Time:        149,
	}

	jb, err := tky.AsSignable()
	assert.NoError(err)

	tky.Sig = ed25519.Sign(sk, jb)

	assert.True(tky.Verify(pk))
}

/*
// ExampleInsecureSend to serve as a reference for a insecure blocking unary
// gRPC request over unix socket
//
// Could be useful in development networks (localnet, devnet) Also useful when
// client co-deployed with the node
func TestExampleInsecureSend(t *testing.T) {

	address := "unix://tmp/dusk-grpc.sock"
	password := "nopass"

	dialCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Set up an attempt (with timeout) to connect to the server
	conn, err := grpc.DialContext(dialCtx, address,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		log.WithError(err).Error("could not connect")
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	client := pb.NewWalletClient(conn)

	// Request timeout param set
	reqCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Request parameters
	req := pb.LoadRequest{Password: password}

	_, err = client.LoadWallet(reqCtx, &req)
	if err != nil {
		log.WithError(err).Error("could not send")
		return
	}
}

// ExampleSecureSend to serve as an example for a blocking unary gRPC request with
// Basic HTTP authorization and TLS enabled over TCP
//
// Could be useful in public networks (testnet,mainnet)
func TestExampleSecureSend(t *testing.T) {

	t.Skip("test requires manual setup")

	address := "127.0.0.1:9000"
	password := "nopass"
	certFile := "/tmp/ca.cert"
	// name use to verify the hostname returned by TLS handshake
	hostname := "www.example.com"

	fmt.Println("Sending request")

	// Create tls based credential.
	creds, err := credentials.NewClientTLSFromFile(certFile, hostname)
	if err != nil {
		log.WithError(err).Fatalf("could not load credentials")
		return
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Set up a connection to the server.
	conn, err := grpc.DialContext(dialCtx, address,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(basicAuth{
			username: "user",
			password: "password",
			secured:  true,
		}))
	if err != nil {
		log.WithError(err).Error("could not connect")
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	client := pb.NewWalletClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := pb.LoadRequest{Password: password}
	_, err = client.LoadWallet(ctx, &req)
	if err != nil {
		log.WithError(err).Error("could not send")
		return
	}

	///Output: Sending request
}

// nolint
type basicAuth struct {
	username string
	password string
	secured  bool
}

func (b basicAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	auth := b.username + ":" + b.password
	enc := base64.StdEncoding.EncodeToString([]byte(auth))
	return map[string]string{
		"authorization": "Basic " + enc,
	}, nil
}

func (b basicAuth) RequireTransportSecurity() bool {
	return b.secured
}
*/
