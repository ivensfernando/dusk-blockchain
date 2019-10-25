package firststep

import (
	"bytes"
	"time"

	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/agreement"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/header"
	"github.com/dusk-network/dusk-blockchain/pkg/core/consensus/reduction"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/encoding"
	"github.com/dusk-network/dusk-blockchain/pkg/p2p/wire/topics"
	"github.com/dusk-network/dusk-blockchain/pkg/util/nativeutils/eventbus"
	"github.com/dusk-network/dusk-blockchain/pkg/util/nativeutils/rpcbus"
	"github.com/dusk-network/dusk-wallet/key"
	log "github.com/sirupsen/logrus"
)

var lg = log.WithField("process", "first-step reduction")
var emptyHash = [32]byte{}
var regenerationPackage = new(bytes.Buffer)

type Reducer struct {
	broker     eventbus.Broker
	rpcBus     *rpcbus.RPCBus
	keys       key.ConsensusKeys
	stepper    consensus.Stepper
	signer     consensus.Signer
	subscriber consensus.Subscriber

	reductionID uint32

	handler    *reduction.Handler
	aggregator *aggregator
	timeOut    time.Duration
	Timer      *reduction.Timer
}

// NewComponent returns an uninitialized reduction component.
func NewComponent(broker eventbus.Broker, rpcBus *rpcbus.RPCBus, keys key.ConsensusKeys, timeOut time.Duration) consensus.Component {
	return &Reducer{
		broker:  broker,
		rpcBus:  rpcBus,
		keys:    keys,
		timeOut: timeOut,
	}
}

// Initialize the reduction component, by instantiating the handler and creating
// the topic subscribers.
// Implements consensus.Component
func (r *Reducer) Initialize(stepper consensus.Stepper, signer consensus.Signer, ru consensus.RoundUpdate) []consensus.TopicListener {
	r.stepper = stepper
	r.signer = signer
	r.handler = reduction.NewHandler(r.keys, ru.P)
	r.Timer = reduction.NewTimer(r.Halt)

	bestScoreSubscriber := consensus.TopicListener{
		Topic:    topics.BestScore,
		Listener: consensus.NewSimpleListener(r.CollectBestScore),
	}

	return []consensus.TopicListener{bestScoreSubscriber}
}

// Finalize the Reducer component by killing the timer, if it is still running.
// This will stop a reduction cycle short, and renders this Reducer useless
// after calling.
func (r *Reducer) Finalize() {
	r.Timer.Stop()
}

func (r *Reducer) CollectReductionEvent(e consensus.Event) error {
	ev := reduction.New()
	if err := reduction.Unmarshal(&e.Payload, ev); err != nil {
		return err
	}

	if err := r.handler.VerifySignature(e.Header, ev.SignedHash); err != nil {
		return err
	}

	r.aggregator.collectVote(*ev, e.Header)
	return nil
}

func (r *Reducer) Filter(hdr header.Header) bool {
	return !r.handler.IsMember(hdr.PubKeyBLS, hdr.Round, hdr.Step)
}

func (r *Reducer) startReduction() {
	r.Timer.Start(r.timeOut)
	r.aggregator = newAggregator(r.Halt, r.handler, r.rpcBus)
}

func (r *Reducer) sendReduction(hash []byte) {
	sig, err := r.signer.Sign(hash, nil)
	if err != nil {
		lg.WithField("category", "BUG").WithError(err).Errorln("error in signing reduction")
		return
	}

	payload := new(bytes.Buffer)
	if err := encoding.WriteBLS(payload, sig); err != nil {
		lg.WithField("category", "BUG").WithError(err).Errorln("error in encoding BLS signature")
		return
	}

	if err := r.signer.SendAuthenticated(topics.Reduction, hash, payload); err != nil {
		lg.WithField("category", "BUG").WithError(err).Errorln("error in sending authenticated Reduction")
	}
}

func (r *Reducer) Halt(hash []byte, svs ...*agreement.StepVotes) {
	r.subscriber.Unsubscribe(r.reductionID)
	buf := new(bytes.Buffer)
	if len(svs) > 0 {
		if err := agreement.MarshalStepVotes(buf, svs[0]); err != nil {
			lg.WithField("category", "BUG").WithError(err).Errorln("error in marshalling StepVotes")
			return
		}
	}

	r.broker.Publish(topics.StepVotes, buf)
	r.stepper.RequestStepUpdate()
}

// CollectBestScore activates the 2-step reduction cycle.
func (r *Reducer) CollectBestScore(e consensus.Event) error {
	listener := consensus.NewFilteringListener(r.CollectReductionEvent, r.Filter)
	r.subscriber.Subscribe(topics.Reduction, listener)
	r.reductionID = listener.ID()

	// sending reduction can very well be done concurrently
	go r.sendReduction(e.Payload.Bytes())
	r.startReduction()

	return nil
}
