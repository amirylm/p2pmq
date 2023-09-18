# DON Composition

`p2pmq` enables to compose a network of peers cross DONs (Decentralized Oracles Network), acting as a decentralized message bus for DON to DON communication.

## Overview

`p2pmq` agents can run as a sidecar to the DON's nodes, and enables to gossip messages over topics with optimal latency, while enabling a decoupled message validation to avoid introducing additional dependencies for the agent e.g. public keys, persistent storage of reports, etc.

Gossiping OCR reports enables to achieve optimal latency and throughput, while maintaining an optimal network topology, w/o external components or ledgers.

Scoring and msg validation are used to protect the network from bad actors and ensure integrity.

The following diagram visualizes the composition of a network of `p2pmq` agents across DONs:

![p2pmq DON Composition](./resources/img/composer-p2pmq.png)

<br />

## Messaging

DONs communication is based on OCR reports, which are broadcasted over some pubsub topic rather than on-chain transmission.

The reports MUST be signed by a quorum of the DON's nodes, otherwise they are considered invalid and nodes that broadcast them are penalized.

### Message Validation

`p2pmq` enables to aid in a custom, decoupled validation before processing 
and propagating messages to the network. 
The validation is done on top of open gRPC duplex stream to ensure 
high throughput and low latency as possible.

The actual validation needs to verify that a given report was originated by some DON, where at least a quorum of nodes have confirmed it. 

**TBD** public key sharing cross DONs.

In addition, sequence number is used to ensure message order and penalize bad actors
that sends unrealistic sequence numbers.

The disincentivation of sending and propogating invalid messages across the network, helps to protect the network from bad actors. Enabling a trustless environment for DONs to communicate.

<br />
