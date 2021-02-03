// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT License was not distributed with this
// file, you can obtain one at https://opensource.org/licenses/MIT.
//
// Copyright (c) DUSK NETWORK. All rights reserved.

package chain

import (
	"testing"

	"github.com/dusk-network/dusk-blockchain/pkg/config"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus"
	"github.com/dusk-network/dusk-blockchain/pkg/core/data/block"
	"github.com/dusk-network/dusk-blockchain/pkg/core/database"
	"github.com/dusk-network/dusk-blockchain/pkg/core/database/lite"
	"github.com/dusk-network/dusk-blockchain/pkg/core/tests/helper"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/topics"
	assert "github.com/stretchr/testify/require"
)

func TestSuccessiveBlocks(t *testing.T) {
	assert := assert.New(t)
	s, c := setupSynchronizerTest()

	// tipHeight will be 0, so make the successive block
	blk := helper.RandomBlock(1, 1)
	res, err := s.processBlock(0, *blk)
	assert.NoError(err)
	assert.Nil(res)

	// Block will come through the catchblock channel
	rBlk := <-c
	assert.True(blk.Equals(&rBlk.Blk))
}

func TestFutureBlocks(t *testing.T) {
	assert := assert.New(t)
	s, _ := setupSynchronizerTest()

	height := uint64(10)
	blk := helper.RandomBlock(height, 1)
	resp, err := s.processBlock(0, *blk)
	assert.NoError(err)

	// Response should be of the GetBlocks topic
	assert.Equal(resp[0].Bytes()[0], uint8(topics.GetBlocks))

	// Block should be in the sequencer
	assert.NotEmpty(s.sequencer.blockPool[height])
}

// TODO: this needs to be moved to the chain.
// func TestSyncProgress(t *testing.T) {
// 	assert := assert.New(t)
// 	s, m := setupSynchronizerTest()
//
// 	// SyncProgress should be 0% right now
// 	resp, err := s.GetSyncProgress(context.Background(), &node.EmptyRequest{})
// 	assert.NoError(err)
//
// 	assert.Equal(resp.Progress, float32(0.0))
//
// 	// Change tipHeight and then give the synchronizer a block from far in the future
// 	m.tipHeight = 50
// 	blk := helper.RandomBlock(100, 1)
// 	s.processBlock(*blk)
//
// 	// SyncProgress should be 50%
// 	resp, err = s.GetSyncProgress(context.Background(), &node.EmptyRequest{})
// 	assert.NoError(err)
//
// 	assert.Equal(resp.Progress, float32(50.0))
// }

func setupSynchronizerTest() (*synchronizer, chan consensus.Results) {
	c := make(chan consensus.Results, 1)
	m := &mockChain{tipHeight: 0, catchBlockChan: c}
	_, db := lite.CreateDBConnection()
	// Give DB a genesis to avoid errors
	genesis := config.DecodeGenesis()

	if err := db.Update(func(t database.Transaction) error {
		return t.StoreBlock(genesis)
	}); err != nil {
		panic(err)
	}

	return newSynchronizer(db, m), c
}

type mockChain struct {
	tipHeight      uint64
	catchBlockChan chan consensus.Results
}

func (m *mockChain) CurrentHeight() uint64 {
	return m.tipHeight
}

func (m *mockChain) ProcessSuccessiveBlock(blk block.Block) error {
	m.catchBlockChan <- consensus.Results{Blk: blk, Err: nil}
	return nil
}

func (m *mockChain) ProcessSyncBlock(_ block.Block) error {
	return nil
}

func (m *mockChain) ProduceBlock() error {
	return nil
}

func (m *mockChain) StopBlockProduction() {}
