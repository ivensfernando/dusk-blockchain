# Agreement component

This package implements the component which performs the [Agreement Phase](./agreement.md) of the consensus.

## Values

### Block Agreement Event

| Field | Type |
| :--- | :--- |
| Aggregated public keys\* | BLS APK |
| Committee bit-representation\* | uint64 |
| Aggregated signatures\* | BLS Signature |
| Signature of all fields | BLS Signature |

\* These fields appear twice, once for each step of [Reduction](../reduction/reduction.md).

## Architecture

The `Agreement` component is implemented within the `Loop` struct, found in `step.go`. This struct contains all the data and logic that the agreement component needs to do its job. Unlike the other consensus components, the `Loop` struct implements `ControlFn`, as the component works slightly differently to the others - it needs to be started concurrently to the other components, and not in sequence.

The agreement component is given access to the DB, and has an instance of a `candidate.Requestor`, which can be used to request certain pending blocks from the network.

### Main component

When the agreement component starts up (with the `Run` function), it flushes all queued agreement events for the given round, and passes them directly to the processing function (`collectEvent`). Afterwards, the component starts listening on three different channels:

- The `agreementChan`, which transports agreement events from the eventbus to the agreement component
- The `CollectedVotesChan`, which collects final results from the [accumulator](#accumulator)
- The `ctx.Done` channel, which notifies the agreement component of context cancellations

Events sent through the `agreementChan` will go immediately to the processing function, `collectEvent`. This sends the event off to the local `Accumulator`.

Once a final result is received through the `CollectedVotesChan`, the agreement component checks the DB to see if a candidate for the resulting block hash exists. If not, it makes use of its own `candidate.Requestor` to fetch it from the network, instead. The winning block is then returned, and the agreement component will stop running.

### Accumulator

For each call to the `Run` function, the agreement component instantiates a new `Accumulator`. This component is responsible for sorting agreement events by step, and keeps track of how many times an event is seen for a given step. Internally, it uses a `handler` which has access to the consensus committee - allowing the `Accumulator` to perform verification, as well as keep track of how close a given step is to **quorum**. Once a step reaches said quorum, the `Accumulator` sends back the collected votes for this step, through the `CollectedVotesChan`, sending it back directly to the main agreement component.

Under the hood, the `Accumulator` makes use of a thread pool in order to parallelize event verification and storage. The amount of workers is configurable, but is standardly set to 4.

To store events, the `Accumulator` uses a `store`, which is a thread-safe wrapper around a map.
