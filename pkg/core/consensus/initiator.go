package consensus

import (
	"bytes"
	"errors"

	"github.com/dusk-network/dusk-blockchain/pkg/core/transactions"
)

// InCommittee will query the blockchain for any non-expired stakes that belong to the supplied public key.
func InCommittee(blsPubKey []byte) bool {
	retriever := NewTxRetriever(nil, FindStake)
	_, err := retriever.SearchForTx(blsPubKey)
	return err != nil
}

func FindStake(txs []transactions.Transaction, item []byte) (transactions.Transaction, error) {
	for _, tx := range txs {
		stake, ok := tx.(*transactions.Stake)
		if !ok {
			continue
		}

		if bytes.Equal(item, stake.PubKeyBLS) {
			return stake, nil
		}
	}

	return nil, errors.New("could not find corresponding stake")
}
