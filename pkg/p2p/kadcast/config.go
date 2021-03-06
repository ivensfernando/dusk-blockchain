// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT License was not distributed with this
// file, you can obtain one at https://opensource.org/licenses/MIT.
//
// Copyright (c) DUSK NETWORK. All rights reserved.

package kadcast

// Default kadcast configuration.
//
// a.k.a globally known parameters determining the redundancy
// Typical parameter values are k ∈ [20, 100] and α = 3.
// TODO: Current impl does not allow α > 1.

// DefaultMaxBucketPeers is the maximum number of peers that a `bucket` can hold.
var DefaultMaxBucketPeers uint8 = 25

// DefaultMaxBetaDelegates maximum number of delegates per bucket.
var DefaultMaxBetaDelegates uint8 = 3

// DefaultAlphaClosestNodes the node looks up the α closest nodes regarding the XOR-metric in its own buckets.
var DefaultAlphaClosestNodes int = 1

// DefaultKNumber is the K number of peers that a node will send on a `FIND_NODES` message.
var DefaultKNumber int = 20

const (

	// MaxTCPacketSize is the max size allowed of TCP packet.
	MaxTCPacketSize = 500000

	// MaxUDPacketSize max size of a UDP packet. As per default MTU 1500, UDP
	// packet should be up to 1472 bytes.
	MaxUDPacketSize = 1472
)
