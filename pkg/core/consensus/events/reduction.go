package events

import (
	"bytes"
	"encoding/binary"

	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/user"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/crypto/bls"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/encoding"
	"golang.org/x/crypto/ed25519"
)

type (
	// Reduction is a basic reduction event.
	Reduction struct {
		*Header
		VotedHash  []byte
		SignedHash []byte
	}

	// ReductionUnMarshaller marshals and unmarshales th
	ReductionUnMarshaller struct {
		*UnMarshaller
	}

	ReductionUnmarshaller interface {
		wire.EventMarshaller
		wire.EventDeserializer
		// HACK: why are these 2 methods here?
		MarshalVoteSet(*bytes.Buffer, []wire.Event) error
		UnmarshalVoteSet(*bytes.Buffer) ([]wire.Event, error)
	}

	// OutgoingReductionUnmarshaller unmarshals a Reduction event ready to be gossipped. Its intent is to centralize the code part that has responsibility of signing it
	OutgoingReductionUnmarshaller struct {
		ReductionUnmarshaller
	}
)

// NewReduction returns and empty Reduction event.
func NewReduction() *Reduction {
	return &Reduction{
		Header: &Header{},
	}
}

// Equal as specified in the Event interface
func (e *Reduction) Equal(ev wire.Event) bool {
	other, ok := ev.(*Reduction)
	return ok && (bytes.Equal(e.PubKeyBLS, other.PubKeyBLS)) &&
		(e.Round == other.Round) && (e.Step == other.Step)
}

func NewReductionUnMarshaller() *ReductionUnMarshaller {
	return &ReductionUnMarshaller{NewUnMarshaller()}
}

func (r *ReductionUnMarshaller) NewEvent() wire.Event {
	return NewReduction()
}

// Unmarshal unmarshals the buffer into a Committee
func (a *ReductionUnMarshaller) Unmarshal(r *bytes.Buffer, ev wire.Event) error {
	bev := ev.(*Reduction)
	if err := a.HeaderUnmarshaller.Unmarshal(r, bev.Header); err != nil {
		return err
	}

	if err := encoding.Read256(r, &bev.VotedHash); err != nil {
		return err
	}

	if err := encoding.ReadBLS(r, &bev.SignedHash); err != nil {
		return err
	}

	return nil
}

// Marshal a Reduction into a buffer.
func (a *ReductionUnMarshaller) Marshal(r *bytes.Buffer, ev wire.Event) error {
	bev := ev.(*Reduction)
	if err := a.HeaderMarshaller.Marshal(r, bev.Header); err != nil {
		return err
	}

	if err := encoding.Write256(r, bev.VotedHash); err != nil {
		return err
	}

	if err := encoding.WriteBLS(r, bev.SignedHash); err != nil {
		return err
	}

	return nil
}

// TODO: generalize the set/array marshalling/unmarshalling through an interface
func (a *ReductionUnMarshaller) UnmarshalVoteSet(r *bytes.Buffer) ([]wire.Event, error) {
	length, err := encoding.ReadVarInt(r)
	if err != nil {
		return nil, err
	}

	evs := make([]wire.Event, length)
	for i := uint64(0); i < length; i++ {
		rev := &Reduction{
			Header: &Header{},
		}
		if err := a.Unmarshal(r, rev); err != nil {
			return nil, err
		}

		evs[i] = rev
	}

	return evs, nil
}

func (a *ReductionUnMarshaller) MarshalVoteSet(r *bytes.Buffer, evs []wire.Event) error {
	if err := encoding.WriteVarInt(r, uint64(len(evs))); err != nil {
		return err
	}

	for _, event := range evs {
		if err := a.Marshal(r, event); err != nil {
			return err
		}
	}

	return nil
}

// SignReduction is a shortcut to BLS and ED25519 sign a reduction message
func SignReduction(buf *bytes.Buffer, keys *user.Keys) error {
	e := NewReduction()
	unMarshaller := NewReductionUnMarshaller()
	if err := unMarshaller.Unmarshal(buf, e); err != nil {
		return err
	}

	if err := SignReductionEvent(e, keys); err != nil {
		return err
	}

	outbuf := new(bytes.Buffer)
	if err := unMarshaller.Marshal(outbuf, e); err != nil {
		return err
	}

	signature := ed25519.Sign(*keys.EdSecretKey, outbuf.Bytes())

	signed := new(bytes.Buffer)
	if err := encoding.Write512(signed, signature); err != nil {
		return err
	}

	if err := encoding.Write256(signed, signature); err != nil {
		return err
	}

	if _, err := signed.Write(outbuf.Bytes()); err != nil {
		return err
	}

	*buf = *signed
	return nil
}

// SignReductionEvent is a shortcut to create a BLS signature of a reduction vote and fill the proper field in Reduction struct
func SignReductionEvent(ev *Reduction, keys *user.Keys) error {
	buf := new(bytes.Buffer)

	//TODO: Vote is actually the Header
	vote := &Vote{}
	vote.Round = ev.Round
	vote.Step = ev.Step
	vote.BlockHash = ev.VotedHash

	if err := MarshalSignableVote(buf, vote); err != nil {
		return err
	}

	signedHash, err := bls.Sign(keys.BLSSecretKey, keys.BLSPubKey, ev.VotedHash)
	if err != nil {
		return err
	}
	ev.SignedHash = signedHash.Compress()
	ev.Header.PubKeyBLS = keys.BLSPubKeyBytes
	return nil
}

func NewOutgoingReductionUnmarshaller() *OutgoingReductionUnmarshaller {
	return &OutgoingReductionUnmarshaller{
		ReductionUnmarshaller: NewReductionUnMarshaller(),
	}
}

func (a *OutgoingReductionUnmarshaller) NewEvent() wire.Event {
	return NewReduction()
}

func (a *OutgoingReductionUnmarshaller) Unmarshal(reductionBuffer *bytes.Buffer, ev wire.Event) error {
	rev := ev.(*Reduction)
	if err := encoding.ReadUint64(reductionBuffer, binary.LittleEndian, &rev.Round); err != nil {
		return err
	}

	if err := encoding.ReadUint8(reductionBuffer, &rev.Step); err != nil {
		return err
	}

	if err := encoding.Read256(reductionBuffer, &rev.VotedHash); err != nil {
		return err
	}

	return nil
}