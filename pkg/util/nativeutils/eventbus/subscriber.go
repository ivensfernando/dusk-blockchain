// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT License was not distributed with this
// file, you can obtain one at https://opensource.org/licenses/MIT.
//
// Copyright (c) DUSK NETWORK. All rights reserved.

package eventbus

import (
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/topics"
	lg "github.com/sirupsen/logrus"
)

// Subscriber subscribes a channel to Event notifications on a specific topic.
type Subscriber interface {
	Subscribe(topic topics.Topic, listener Listener) uint32
	Unsubscribe(topics.Topic, uint32)
}

// Subscribe subscribes to a topic with a channel.
func (bus *EventBus) Subscribe(topic topics.Topic, listener Listener) uint32 {
	return bus.listeners.Store(topic, listener)
}

// Unsubscribe removes all listeners defined for a topic.
func (bus *EventBus) Unsubscribe(topic topics.Topic, id uint32) {
	found := bus.listeners.Delete(topic, id)

	logEB.WithFields(lg.Fields{
		"found": found,
		"topic": topic,
	}).Traceln("unsubscribing")
}
