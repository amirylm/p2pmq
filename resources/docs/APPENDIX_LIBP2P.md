# Appendix: Libp2p

## Links

- [Libp2p specs](https://github.com/libp2p/specs)
- [Gossipsub v1.1 spec](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.1.md)
- [Gossipsub v1.1 Evaluation Report](https://gateway.ipfs.io/ipfs/QmRAFP5DBnvNjdYSbWhEhVRJJDFCLpPyvew5GwCCB4VxM4)

## Overview

Libp2p is a modular networking framework designed for peer-to-peer communication in decentralized systems. It provides a foundation for building decentralized applications and systems by offering a range of essential components.

Libp2p was chosen because it provides a battle tested, complete yet extensible networking framework for a distributed message engine.

## Libp2p Protocols

The following libp2p protocols are utilized by the DME:

| Name | Description | Links |
| --- | --- | --- |
| Gossipsub (v1.1) | Pubsub messaging | https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.1.md |
| Kad DHT | Distributed hash table for discovery | https://github.com/libp2p/specs/tree/master/kad-dht |
| Noise | Protocol for securing transport | https://github.com/libp2p/specs/blob/master/noise/README.md |
| Yamux | Protocol for multiplexing transport | https://github.com/libp2p/specs/blob/master/yamux/README.md |

## Security

Libp2p provides strong cryptographic support, including secure key management and encryption. It offers options for transport-level encryption and authentication, ensuring the confidentiality and integrity of data exchanged between peers.

Secure channels in libp2p are established with the help of a transport upgrader, which providers layers of security and stream multiplexing over "raw" connections such as TCP sockets.

## Kad DHT

[KadDHT](https://github.com/libp2p/specs/tree/master/kad-dht) is used for peer discovery. It requires to deploy a set of bootstrapper nodes, which are used by new peers to join the network.

**NOTE:** bootstrappers should managed by multiple parties for decentralization.

## Pubsub

Libp2p's pubsub system is designed to be extensible by more specialized routers and provides an optimized environment for a distributed protocol that runs over a trustless p2p network.

### Key Concepts of Gossiping

Gossipsub v1.1 is a dynamic and efficient message propagation protocol, it is based on randomized topic meshes and gossip, with moderate amplification factors and good scaling properties.

* Gossipsub introduces a **heartbeat** mechanism that defines a regular interval at which peers exchange heartbeat messages with their mesh peers
    * Heartbeats help maintain the freshness of the mesh and enable the discovery of new topics or subscriptions
    * During heartbeats, peers can also inform each other about their latest state, including the topics they are interested in
* Gossipsub orchestrates dynamic, real-time subscription/peering management while outsourcing peer discovery for flexibility
* Gossipsub facilitates concurrent & configurable validation system that allows to outsource validation using [extended validators](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.1.md#extended-validators)
* Gossipsub incorporates a **peer scoring** mechanism to evaluate the behavior and trustworthiness of network peers. Peers are assigned scores based on various factors, including their message forwarding reliability, responsiveness, and adherence to protocol rules
    * Low-scoring peers may be disconnected from the network to maintain its health and reliability
* Gossiping enables to have a loosely connected network of nodes, while all-2-all requires fully connected network
    * helps to scale and overcome bad network conditions where we don't have connectivity among all participants
* Gossipsub is using a msg_id concept which is useful for de-duplication, to ensure each message is processed only once

### How Gossipsub Works

- **Message Exchange:** When a node wishes to publish a message to a topic, it first sends the message to its mesh peers within the corresponding fanout group. Mesh peers validate the message and, if valid, forward it to their mesh peers and so on. Messages propagate through the mesh until they reach all mesh peers within the group.
- **Heartbeat:** Each peer runs a periodic stabilization process called the "heartbeat procedure"
at regular intervals (1s is the default). The heartbeat serves three functions: 
1. [mesh maintenance](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.0.md#mesh-maintenance) to keep mesh fresh and discover new subscriptions
2. [fanout maintenance](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.0.md#fanout-maintenance) to efficiently adjust fanout groups based on changing interests
3. [gossip emission](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.0.md#gossip-emission) to a random selection of peers for each topic (that are not already members of the topic mesh)
- **Piggybacking:** Gossipsub employs [piggybacking](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.0.md#control-message-piggybacking) - a technique where acknowledgment (ACK) messages carry new messages. When a peer sends an ACK for a message, it may include new messages it has received. This optimizes bandwidth usage by combining acknowledgment and message propagation in a single step.

