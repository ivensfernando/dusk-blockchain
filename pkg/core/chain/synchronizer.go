package chain

import (
	"bytes"
	"context"

	"github.com/dusk-network/dusk-blockchain/pkg/config"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/block"
	"github.com/dusk-network/dusk-blockchain/pkg/core/database"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/message"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/topics"
	"github.com/dusk-network/dusk-blockchain/pkg/util/nativeutils/eventbus"
	"github.com/dusk-network/dusk-blockchain/pkg/util/nativeutils/rpcbus"
	"github.com/dusk-network/dusk-protobuf/autogen/go/node"
)

// Synchronizer acts as the gateway for incoming blocks from the network.
// It decides how the Chain should process these blocks, and is responsible
// for requesting missing items in case of a desync.
type Synchronizer struct {
	eb eventbus.Broker
	rb *rpcbus.RPCBus
	db database.DB

	highestSeen uint64
	syncTarget  uint64
	*sequencer

	state syncState

	ctx context.Context

	chain Ledger
}

type syncState func(currentHeight uint64, blk block.Block) (syncState, []bytes.Buffer, error)

func (s *Synchronizer) inSync(currentHeight uint64, blk block.Block) (syncState, []bytes.Buffer, error) {
	if blk.Header.Height > currentHeight+1 {
		// If this block is from far in the future, we should start syncing mode.
		s.sequencer.add(blk)
		s.chain.StopBlockProduction()
		b, err := s.startSync(blk.Header.Height, currentHeight)
		return s.outSync, b, err
	}

	// otherwise notify the chain (and the consensus loop)
	s.chain.ProcessSucceedingBlock(blk)
	return s.inSync, nil, nil
}

func (s *Synchronizer) outSync(currentHeight uint64, blk block.Block) (syncState, []bytes.Buffer, error) {
	if blk.Header.Height > currentHeight+1 {
		// if there is a gap we add the future block to the sequencer
		s.sequencer.add(blk)
		return s.outSync, nil, nil
	}

	// Retrieve all successive blocks that need to be accepted
	blks := s.sequencer.provideSuccessors(blk)

	for _, blk := range blks {
		// append them all to the ledger
		if err := s.chain.ProcessSyncBlock(blk); err != nil {
			log.WithError(err).Debug("could not AcceptBlock")
			return s.outSync, nil, err
		}

		if blk.Header.Height == s.syncTarget {
			// if we reach the target we get into sync mode
			// and trigger the consensus again
			go func() {
				if err := s.chain.ProduceBlock(s.ctx); err != nil {
					// TODO we need to have a recovery procedure rather than
					// just log and forget
					log.WithError(err).Error("crunchBlocks exited with error")
				}
			}()
			return s.inSync, nil, nil
		}
	}

	return s.outSync, nil, nil
}

// NewSynchronizer returns an initialized Synchronizer, ready for use.
func NewSynchronizer(ctx context.Context, eb eventbus.Broker, rb *rpcbus.RPCBus, db database.DB, chain Ledger) *Synchronizer {
	s := &Synchronizer{
		eb:        eb,
		rb:        rb,
		db:        db,
		sequencer: newSequencer(),
		ctx:       ctx,
		chain:     chain,
	}
	s.state = s.inSync
	return s
}

// ProcessBlock handles an incoming block from the network.
func (s *Synchronizer) ProcessBlock(m message.Message) (res []bytes.Buffer, err error) {
	blk := m.Payload().(block.Block)

	currentHeight := s.chain.CurrentHeight()

	// Is it worth looking at this?
	if blk.Header.Height <= currentHeight {
		log.Debug("discarded block from the past")
		return
	}

	s.state, res, err = s.state(currentHeight, blk)
	return
}

func (s *Synchronizer) startSync(tipHeight, currentHeight uint64) ([]bytes.Buffer, error) {
	s.syncTarget = tipHeight
	if s.syncTarget > currentHeight+config.MaxInvBlocks {
		s.syncTarget = currentHeight + config.MaxInvBlocks
	}

	var hash []byte
	if err := s.db.View(func(t database.Transaction) error {
		var err error
		hash, err = t.FetchBlockHashByHeight(currentHeight)
		return err
	}); err != nil {
		return nil, err
	}

	msgGetBlocks := createGetBlocksMsg(hash)
	return marshalGetBlocks(msgGetBlocks)
}

// GetSyncProgress returns how close the node is to being synced to the tip,
// as a percentage value.
func (s *Synchronizer) GetSyncProgress(ctx context.Context, e *node.EmptyRequest) (*node.SyncProgressResponse, error) {
	if s.highestSeen == 0 {
		return &node.SyncProgressResponse{Progress: 0}, nil
	}

	prevBlockHeight := s.chain.CurrentHeight()
	progressPercentage := (float64(prevBlockHeight) / float64(s.highestSeen)) * 100

	// Avoiding strange output when the chain can be ahead of the highest
	// seen block, as in most cases, consensus terminates before we see
	// the new block from other peers.
	if progressPercentage > 100 {
		progressPercentage = 100
	}

	return &node.SyncProgressResponse{Progress: float32(progressPercentage)}, nil
}

func createGetBlocksMsg(latestHash []byte) *message.GetBlocks {
	msg := &message.GetBlocks{}
	msg.Locators = append(msg.Locators, latestHash)
	return msg
}

//nolint:unparam
func marshalGetBlocks(msg *message.GetBlocks) ([]bytes.Buffer, error) {
	buf := topics.GetBlocks.ToBuffer()
	if err := msg.Encode(&buf); err != nil {
		//FIXME: shall this panic here ?  result 1 (error) is always nil (unparam)
		//log.Panic(err)
		return nil, err
	}

	return []bytes.Buffer{buf}, nil
}
