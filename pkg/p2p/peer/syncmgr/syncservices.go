package syncmgr

import (
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/blockchain"
	"gitlab.dusk.network/dusk-core/dusk-go/pkg/p2p/wire/payload"
)

func getHeaders(chain blockchain.Chain, msg *payload.MsgGetHeaders) (*payload.MsgHeaders, error) {
	var msgheaders payload.MsgHeaders
	locator := msg.Locator
	hashStop := msg.HashStop

	headers, err := chain.GetHeaders(locator, hashStop)
	if err != nil {
		return nil, err
	}

	msgheaders.Headers = headers
	return &msgheaders, nil
}