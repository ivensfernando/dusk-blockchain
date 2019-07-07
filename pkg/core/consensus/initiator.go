package consensus

import (
	"bytes"

	cfg "gitlab.dusk.network/dusk-core/dusk-go/pkg/config"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/block"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/consensus/user"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/database"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/core/transactions"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/protocol"
)

func GetStartingRound(eventBroker wire.EventBroker, db database.DB, keys user.Keys) (uint64, error) {
	// Get a db connection
	if db == nil {
		drvr, err := database.From(cfg.Get().Database.Driver)
		if err != nil {
			return 0, err
		}

		db, err = drvr.Open(cfg.Get().Database.Dir, protocol.MagicFromConfig(), false)
		if err != nil {
			return 0, err
		}
	}

	found := findActiveStakes(keys, getCurrentHeight(db), db)

	// Start listening for accepted blocks, regardless of if we found stakes or not
	acceptedBlockChan, listener := InitAcceptedBlockUpdate(eventBroker)

	// Unsubscribe from AcceptedBlock once we're done
	defer listener.Quit()

	for {
		blk := <-acceptedBlockChan
		if found || keyFound(keys, blk.Txs) {
			return blk.Header.Height + 1, nil
		}
	}
}

func getCurrentHeight(db database.DB) uint64 {
	var height uint64
	err := db.View(func(t database.Transaction) error {
		state, err := t.FetchState()
		if err != nil {
			return err
		}

		header, err := t.FetchBlockHeader(state.TipHash)
		if err != nil {
			return err
		}

		height = header.Height
		return nil
	})

	if err != nil {
		return 0
	}

	return height
}

func findActiveStakes(keys user.Keys, currentHeight uint64, db database.DB) bool {
	searchingHeight := currentHeight - transactions.MaxLockTime
	if currentHeight < transactions.MaxLockTime {
		searchingHeight = 0
	}

	for {
		var b *block.Block
		err := db.View(func(t database.Transaction) error {
			hash, err := t.FetchBlockHashByHeight(searchingHeight)
			if err != nil {
				return err
			}

			b, err = t.FetchBlock(hash)
			return err
		})

		if err != nil {
			break
		}

		if keyFound(keys, b.Txs) {
			return true
		}

		searchingHeight++
	}

	return false
}

func keyFound(keys user.Keys, txs []transactions.Transaction) bool {
	for _, tx := range txs {
		stake, ok := tx.(*transactions.Stake)
		if !ok {
			continue
		}

		if bytes.Equal(keys.BLSPubKeyBytes, stake.PubKeyBLS) {
			return true
		}
	}

	return false
}
