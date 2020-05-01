package transactions

import (
	"bytes"
	"encoding/binary"

	"github.com/dusk-network/dusk-blockchain/pkg/core/data/key"

	"github.com/bwesterb/go-ristretto"
)

// Output of a transaction
type Output struct {
	// Commitment to the amount and the mask value
	// This will be generated by the rangeproof
	Commitment ristretto.Point
	amount     ristretto.Scalar
	mask       ristretto.Scalar

	// PubKey refers to the destination key of the receiver
	// One-time pubkey of the receiver
	// Each input will contain a one-time pubkey
	// Only the private key assosciated with this
	// public key can unlock the funds available at this utxo
	PubKey key.StealthAddress

	// Index denotes the position that this output is in the
	// transaction. This is different to the Offset which denotes the
	// position that this output is in, from the start from the blockchain
	Index uint32

	viewKey         key.PublicView
	EncryptedAmount ristretto.Scalar
	EncryptedMask   ristretto.Scalar
}

// NewOutput creates a new transaction Output
func NewOutput(r, amount ristretto.Scalar, index uint32, pubKey key.PublicKey) *Output {
	output := &Output{
		amount:  amount,
		Index:   index,
		PubKey:  *pubKey.StealthAddress(r, index),
		viewKey: *pubKey.PubView,
	}

	return output
}

// EncryptAmount encAmount = amount + H(H(H(r*PubViewKey || index)))
func EncryptAmount(amount, r ristretto.Scalar, index uint32, pubViewKey key.PublicView) ristretto.Scalar {
	rView := pubViewKey.ScalarMult(r)

	rViewIndex := append(rView.Bytes(), uint32ToBytes(index)...)

	var encryptKey ristretto.Scalar
	encryptKey.Derive(rViewIndex)
	encryptKey.Derive(encryptKey.Bytes())
	encryptKey.Derive(encryptKey.Bytes())

	var encryptedAmount ristretto.Scalar
	encryptedAmount.Add(&amount, &encryptKey)

	return encryptedAmount
}

// DecryptAmount decAmount = EncAmount - H(H(H(R*PrivViewKey || index)))
func DecryptAmount(encAmount ristretto.Scalar, R ristretto.Point, index uint32, privViewKey key.PrivateView) ristretto.Scalar {

	var Rview ristretto.Point
	pv := (ristretto.Scalar)(privViewKey)
	Rview.ScalarMult(&R, &pv)

	rViewIndex := append(Rview.Bytes(), uint32ToBytes(index)...)

	var encryptKey ristretto.Scalar
	encryptKey.Derive(rViewIndex)
	encryptKey.Derive(encryptKey.Bytes())
	encryptKey.Derive(encryptKey.Bytes())

	var decryptedAmount ristretto.Scalar
	decryptedAmount.Sub(&encAmount, &encryptKey)

	return decryptedAmount
}

// EncryptMask encMask = mask + H(H(r*PubViewKey || index))
func EncryptMask(mask, r ristretto.Scalar, index uint32, pubViewKey key.PublicView) ristretto.Scalar {
	rView := pubViewKey.ScalarMult(r)
	rViewIndex := append(rView.Bytes(), uint32ToBytes(index)...)

	var encryptKey ristretto.Scalar
	encryptKey.Derive(rViewIndex)
	encryptKey.Derive(encryptKey.Bytes())

	var encryptedMask ristretto.Scalar
	encryptedMask.Add(&mask, &encryptKey)

	return encryptedMask
}

// DecryptMask decMask = Encmask - H(H(r*PubViewKey || index))
func DecryptMask(encMask ristretto.Scalar, R ristretto.Point, index uint32, privViewKey key.PrivateView) ristretto.Scalar {
	var Rview ristretto.Point
	pv := (ristretto.Scalar)(privViewKey)
	Rview.ScalarMult(&R, &pv)

	rViewIndex := append(Rview.Bytes(), uint32ToBytes(index)...)

	var encryptKey ristretto.Scalar
	encryptKey.Derive(rViewIndex)
	encryptKey.Derive(encryptKey.Bytes())

	var decryptedMask ristretto.Scalar
	decryptedMask.Sub(&encMask, &encryptKey)

	return decryptedMask
}

func uint32ToBytes(x uint32) []byte {
	a := make([]byte, 4)
	binary.BigEndian.PutUint32(a, x)
	return a
}

// Equals returns true if two outputs are the same
func (o *Output) Equals(out *Output) bool {
	if o == nil || out == nil {
		return false
	}

	if !bytes.Equal(o.Commitment.Bytes(), out.Commitment.Bytes()) {
		return false
	}

	if !bytes.Equal(o.PubKey.P.Bytes(), out.PubKey.P.Bytes()) {
		return false
	}

	if !bytes.Equal(o.EncryptedAmount.Bytes(), out.EncryptedAmount.Bytes()) {
		return false
	}

	return bytes.Equal(o.EncryptedMask.Bytes(), out.EncryptedMask.Bytes())
}

func marshalOutput(b *bytes.Buffer, o *Output) error {
	if err := binary.Write(b, binary.BigEndian, o.Commitment.Bytes()); err != nil {
		return err
	}

	if err := binary.Write(b, binary.BigEndian, o.PubKey.P.Bytes()); err != nil {
		return err
	}

	if err := binary.Write(b, binary.BigEndian, o.EncryptedAmount.Bytes()); err != nil {
		return err
	}

	if err := binary.Write(b, binary.BigEndian, o.EncryptedMask.Bytes()); err != nil {
		return err
	}

	return nil
}