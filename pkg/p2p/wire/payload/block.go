package payload

import (
	"errors"
	"io"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/crypto/hash"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/encoding"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/crypto/merkletree"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/payload/transactions"
)

// Block defines a block on the Dusk blockchain.
type Block struct {
	Header *BlockHeader
	Txs    []merkletree.Payload
}

// NewBlock will return an empty Block with an empty BlockHeader.
func NewBlock() *Block {
	return &Block{
		Header: &BlockHeader{
			// CertImage should take up space from creation to
			// ensure proper decoding during block selection.
			CertImage: make([]byte, 32),
		},
	}
}

// SetPrevBlock will set all the fields of the Block struct that are
// taken from the previous block.
func (b *Block) SetPrevBlock(prevBlock *Block) error {
	b.Header.Height = prevBlock.Header.Height + 1
	b.Header.PrevBlock = prevBlock.Header.Hash

	// Remove when BLS code is completed
	seedHash, err := hash.Sha3256(prevBlock.Header.Seed)
	if err != nil {
		return err
	}

	b.Header.Seed = seedHash
	return nil
}

// SetSeed will sign the previous block seed with a BLS signature and
// put it in the block. Implement when BLS code is completed
/*func (b *Block) SetSeed(prevSeed []byte, sk *bls.SecretKey) error {
	sig, err := bls.Sign(sk, prevSeed)
	if err != nil {
		return err
	}

	b.Header.Seed = sig
	return nil
}*/

// SetTime will set the block timestamp.
func (b *Block) SetTime(time int64) {
	b.Header.Timestamp = time
}

// SetRoot will set the block merkle root hash.
func (b *Block) SetRoot() error {
	tree, err := merkletree.NewTree(b.Txs)
	if err != nil {
		return err
	}

	b.Header.TxRoot = tree.MerkleRoot
	return nil
}

// AddTx will add a transaction to the block.
func (b *Block) AddTx(tx *transactions.Stealth) {
	b.Txs = append(b.Txs, tx)
}

// AddCertImage will take a hash from a Certificate and put
// it in the block's CertImage field.
func (b *Block) AddCertImage(cert *Certificate) error {
	if cert.Hash == nil {
		if err := cert.SetHash(); err != nil {
			return err
		}
	}

	b.Header.CertImage = cert.Hash
	return nil
}

// Clear will empty out all the block's fields.
func (b *Block) Clear() {
	b.Header = &BlockHeader{}
	b.Txs = nil
}

// SetHash will set the block hash.
func (b *Block) SetHash() error {
	return b.Header.SetHash()
}

// Encode a Block struct and write to w.
func (b *Block) Encode(w io.Writer) error {
	if err := b.Header.Encode(w); err != nil {
		return err
	}

	lTxs := uint64(len(b.Txs))
	if err := encoding.WriteVarInt(w, lTxs); err != nil {
		return err
	}

	for _, v := range b.Txs {
		tx, ok := v.(*transactions.Stealth)
		if !ok {
			return errors.New("non-tx object found in block txs array")
		}

		if err := tx.Encode(w); err != nil {
			return err
		}
	}

	return nil
}

// Decode a Block struct from r into b.
func (b *Block) Decode(r io.Reader) error {
	b.Header = &BlockHeader{}
	if err := b.Header.Decode(r); err != nil {
		return err
	}

	lTxs, err := encoding.ReadVarInt(r)
	if err != nil {
		return err
	}

	b.Txs = make([]merkletree.Payload, lTxs)
	for i := uint64(0); i < lTxs; i++ {
		tx := &transactions.Stealth{}
		if err := tx.Decode(r); err != nil {
			return err
		}

		b.Txs[i] = tx
	}

	return nil
}