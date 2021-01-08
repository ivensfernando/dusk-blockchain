package candidate

/*
const walletsAmount = 120

// NOTE: exclusively for generating devnet genesis blocks. Generates up to
// 120 wallets, credits all of them with a coinbase output of 50,000 DUSK,
// and gives one of them an initial bid and stake, to allow for network
// bootstrapping.
// NOTE: the RUSK server should be on when running this test.
func TestGenerateDevNetGenesis(t *testing.T) {
	rpcBus := rpcbus.New()

	provideMempoolTxs(rpcBus)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	keyMaster, _ := client.CreateKeysClient(ctx, "127.0.0.1:10000")

	wallets := make([]*wallet.Wallet, walletsAmount)
	for i := 0; i < walletsAmount; i++ {
		w := generateWallet(t, i, keyMaster)
		wallets[i] = w
	}

	// Generate a new genesis block with new wallet pubkey
	g := &generator{
		Emitter:   &consensus.Emitter{RPCBus: rpcBus},
		genPubKey: &wallets[0].PublicKey,
	}

	// TODO: do we need to generate correct proof and score
	seed, _ := crypto.RandEntropy(33)
	proof, _ := crypto.RandEntropy(32)
	score, _ := crypto.RandEntropy(32)

	b, err := g.GenerateBlock(0, seed, proof, score, make([]byte, 32), [][]byte{{0}})
	if err != nil {
		t.Fatal(err)
	}

	// Wallet 0 gets to bootstrap the network
	buf := new(bytes.Buffer)
	if err := encoding.WriteUint64LE(buf, 250000); err != nil {
		t.Fatal(err)
	}

	if err := encoding.WriteVarBytes(buf, wallets[0].Keys().BLSPubKeyBytes); err != nil {
		t.Fatal(err)
	}

	stake := transactions.NewTransaction()
	stake.TxPayload.CallData = buf.Bytes()
	amount := 10000 * wallet.DUSK
	amountBytes := make([]byte, 32)
	binary.LittleEndian.PutUint64(amountBytes[24:32], amount)
	stake.TxPayload.Notes = append(stake.TxPayload.Notes, &transactions.Note{
		Randomness:    &common.JubJubCompressed{Data: make([]byte, 32)},
		PkR:           &common.JubJubCompressed{Data: wallets[0].PublicKey.AG.Data},
		Commitment:    &common.JubJubCompressed{Data: amountBytes},
		Nonce:         &common.BlsScalar{Data: make([]byte, 32)},
		EncryptedData: &common.PoseidonCipher{Data: make([]byte, 96)},
	})

	stake.TxType = transactions.Stake
	b.AddTx(stake)

	buf = new(bytes.Buffer)
	if err := encoding.WriteUint64LE(buf, 250000); err != nil {
		t.Fatal(err)
	}

	m, _ := crypto.RandEntropy(32)
	if err := encoding.Write256(buf, m); err != nil {
		t.Fatal(err)
	}

	bid := transactions.NewTransaction()
	bid.TxPayload.CallData = buf.Bytes()
	bid.TxPayload.Notes = append(bid.TxPayload.Notes, &transactions.Note{
		Randomness:    &common.JubJubCompressed{Data: make([]byte, 32)},
		PkR:           &common.JubJubCompressed{Data: wallets[0].PublicKey.AG.Data},
		Commitment:    &common.JubJubCompressed{Data: amountBytes},
		Nonce:         &common.BlsScalar{Data: make([]byte, 32)},
		EncryptedData: &common.PoseidonCipher{Data: make([]byte, 96)},
	})
	bid.TxType = transactions.Bid
	b.AddTx(bid)

	for _, w := range wallets {
		buf = new(bytes.Buffer)
		if err := encoding.WriteUint64LE(buf, 50000*wallet.DUSK); err != nil {
			t.Fatal(err)
		}

		coinbase := transactions.NewTransaction()
		coinbase.TxPayload.Notes = append(coinbase.TxPayload.Notes, &transactions.Note{
			Randomness:    &common.JubJubCompressed{Data: make([]byte, 32)},
			PkR:           &common.JubJubCompressed{Data: w.PublicKey.AG.Data},
			Commitment:    &common.JubJubCompressed{Data: amountBytes},
			Nonce:         &common.BlsScalar{Data: make([]byte, 32)},
			EncryptedData: &common.PoseidonCipher{Data: make([]byte, 96)},
		})
		coinbase.TxType = transactions.Distribute
		b.AddTx(coinbase)
	}

	// Print the block hex
	buf = new(bytes.Buffer)
	if err := message.MarshalBlock(buf, b); err != nil {
		t.Fatal(err)
	}

	fmt.Println(hex.EncodeToString(buf.Bytes()))
}

func generateWallet(t *testing.T, i int, keyMaster rusk.KeysClient) *wallet.Wallet {
	seed, err := wallet.GenerateNewSeed(nil)
	if err != nil {
		t.Fatal(err)
	}

	db, err := walletdb.New("walletDB" + strconv.Itoa(i))
	if err != nil {
		t.Fatal(err)
	}

	gskr := new(rusk.GenerateKeysRequest)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sk := keys.NewSecretKey()
	pk := keys.NewPublicKey()
	vk := keys.NewViewKey()
	res, err := keyMaster.GenerateKeys(ctx, gskr)
	if err != nil {
		t.Fatal(err)
	}
	keys.USecretKey(res.Sk, sk)
	keys.UPublicKey(res.Pk, pk)
	keys.UViewKey(res.Vk, vk)

	skBuf := new(bytes.Buffer)
	if err = keys.MarshalSecretKey(skBuf, sk); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}

	keysJSON := wallet.KeysJSON{
		Seed:      seed,
		SecretKey: skBuf.Bytes(),
		PublicKey: *pk,
		ViewKey:   *vk,
	}

	w, err := wallet.LoadFromSeed(byte(2), db, "password", "wallet"+strconv.Itoa(i)+".dat", keysJSON)
	if err != nil {
		_ = db.Close()
		t.Fatal(err)
	}

	_ = db.Close()
	if err := os.RemoveAll("walletDB" + strconv.Itoa(i)); err != nil {
		t.Fatal(err)
	}

	return w
}
*/