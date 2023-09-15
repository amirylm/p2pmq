# DON Composition

`p2pmq` enables to compose a network of peers cross DONs (Decentralized Oracles Network), acting as a decentralized message bus for DON to DON communication.

The following diagram illustrates the composition of a network of `p2pmq` peers across DONs:

![p2pmq DON Composition](./resources/img/composer-p2pmq.png)

<br />

## Messaging

DONs communication is based on OCR reports, which are broadcasted over some topic rather than on-chain transmission.

The reports MUST be signed by a quorum of the DON's nodes, otherwise they are considered invalid and any nodes that broadcast them are penalized.

**NOTE** `p2pmq` enables to aid in a custom validation before processing and propagating messages to the network.

**TBD** signature validation cross DONs.

<br />

