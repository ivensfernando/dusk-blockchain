// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT License was not distributed with this
// file, you can obtain one at https://opensource.org/licenses/MIT.
//
// Copyright (c) DUSK NETWORK. All rights reserved.

package agreement

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/committee"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/header"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/key"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/msg"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/user"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/message"
	"github.com/dusk-network/dusk-blockchain/pkg/util/nativeutils/sortedset"
	"github.com/dusk-network/dusk-crypto/bls"
)

// MaxCommitteeSize represents the maximum size of the committee for an
// Agreement quorum.
const MaxCommitteeSize = 64

// Handler interface is handy for tests.
type Handler interface {
	AmMember(uint64, uint8) bool
	IsMember([]byte, uint64, uint8) bool
	Committee(uint64, uint8) user.VotingCommittee
	Quorum(uint64) int
	VotesFor([]byte, uint64, uint8) int
	Verify(message.Agreement) error
}

type handler struct {
	*committee.Handler
}

// NewHandler returns an initialized handler.
//nolint:golint
func NewHandler(keys key.Keys, p user.Provisioners) *handler {
	return &handler{
		Handler: committee.NewHandler(keys, p),
	}
}

// AmMember checks if we are part of the committee.
func (a *handler) AmMember(round uint64, step uint8) bool {
	return a.Handler.AmMember(round, step, MaxCommitteeSize)
}

// IsMember delegates the committee.Handler to check if a Provisioner is in the
// committee for a specified round and step.
func (a *handler) IsMember(pubKeyBLS []byte, round uint64, step uint8) bool {
	return a.Handler.IsMember(pubKeyBLS, round, step, MaxCommitteeSize)
}

// Committee returns a VotingCommittee for a given round and step.
func (a *handler) Committee(round uint64, step uint8) user.VotingCommittee {
	return a.Handler.Committee(round, step, MaxCommitteeSize)
}

// VotesFor delegates embedded committee.Handler to accumulate a vote for a
// given round.
func (a *handler) VotesFor(pubKeyBLS []byte, round uint64, step uint8) int {
	return a.Handler.VotesFor(pubKeyBLS, round, step, MaxCommitteeSize)
}

// Quorum returns the amount of committee members necessary to reach a quorum.
func (a *handler) Quorum(round uint64) int {
	return int(math.Ceil(float64(a.CommitteeSize(round, MaxCommitteeSize)) * 0.75))
}

// Verify checks the signature of the set.
func (a *handler) Verify(ev message.Agreement) error {
	hdr := ev.State()

	if err := verifyWhole(ev); err != nil {
		return fmt.Errorf("failed to verify Agreement Sender: %w", err)
	}

	allVoters := 0

	for i, votes := range ev.VotesPerStep {
		// the beginning step is the same of the second reduction. Since the
		// consensus steps start at 1, this is always a multiple of 3
		// The first reduction step is one less
		step := hdr.Step - 1 + uint8(i)

		// FIXME: what shall we do when step overflows uint8 ?
		if step == math.MaxInt8 {
			err := errors.New("verify, step reached max limit")
			lg.WithError(err).Error("step overflow")

			return err
		}

		committee := a.Committee(hdr.Round, step)
		subcommittee := committee.IntersectCluster(votes.BitSet)

		allVoters += subcommittee.TotalOccurrences()

		apk, err := ReconstructApk(subcommittee.Set)
		if err != nil {
			return fmt.Errorf("failed to reconstruct APK in the Agreement verification: %w", err)
		}

		if err := header.VerifySignatures(hdr.Round, step, hdr.BlockHash, apk, votes.Signature); err != nil {
			return fmt.Errorf("failed to verify BLS multisig: %w", err)
		}
	}

	if allVoters < a.Quorum(hdr.Round) {
		return fmt.Errorf("vote set too small - %v/%v", allVoters, a.Quorum(hdr.Round))
	}

	return nil
}

func (a *handler) getVoterKeys(ev message.Agreement) ([][]byte, error) {
	hdr := ev.State()
	keys := make([][]byte, 0)

	for i, votes := range ev.VotesPerStep {
		step := hdr.Step - 2 + uint8(i)

		// FIXME: what shall we do when step overflows uint8 ?
		if step >= math.MaxInt8 {
			err := errors.New("getVoterKeys, step reached max limit")
			lg.WithError(err).Error("step overflow")

			return nil, err
		}

		committee := a.Committee(hdr.Round, step)
		subcommittee := committee.IntersectCluster(votes.BitSet)

		keys = append(keys, subcommittee.Unravel()...)
	}

	return keys, nil
}

func verifyWhole(a message.Agreement) error {
	hdr := a.State()

	r := new(bytes.Buffer)
	if err := header.MarshalSignableVote(r, hdr); err != nil {
		return err
	}

	// we make a copy of the signature because the crypto package apparently mutates the byte array when
	// Compressing/Decompressing a point
	// see https://github.com/dusk-network/dusk-crypto/issues/16
	sig := make([]byte, len(a.SignedVotes()))
	copy(sig, a.SignedVotes())

	return msg.VerifyBLSSignature(hdr.PubKeyBLS, r.Bytes(), sig)
}

// ReconstructApk reconstructs an aggregated BLS public key from a subcommittee.
func ReconstructApk(subcommittee sortedset.Set) (*bls.Apk, error) {
	var apk *bls.Apk

	if len(subcommittee) == 0 {
		return nil, errors.New("Subcommittee is empty")
	}

	for i, ipk := range subcommittee {
		pk, err := bls.UnmarshalPk(ipk.Bytes())
		if err != nil {
			return nil, err
		}

		if i == 0 {
			apk = bls.NewApk(pk)
			continue
		}

		if err := apk.Aggregate(pk); err != nil {
			return nil, err
		}
	}

	return apk, nil
}
