package agreement

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/crypto"
)

func TestVoteVerification(t *testing.T) {
	// mocking voters
	c, keys := mockCommittee(2, true, 2)
	hash, _ := crypto.RandEntropy(32)
	ev := MockAggregatedAgreementEvent(hash, 1, 2, keys)
	handler := newHandler(c)
	assert.NoError(t, handler.Verify(ev))
}
