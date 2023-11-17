# BLS example

This package shows a high level integration with BLS based networks, using [herumi/bls-eth-go-binary](https://github.com/herumi/bls-eth-go-binary) package for BLS cryptography.

## Links

- [BLS Multi-Signatures With Public-Key Aggregation](https://crypto.stanford.edu/~dabo/pubs/papers/BLSmultisig.html)
- [BLS Signatures Part 1 — Overview](https://alonmuroch-65570.medium.com/bls-signatures-part-1-overview-47d9eebf1c75)
- [BLS signatures in Solidity](https://ethresear.ch/t/bls-signatures-in-solidity/7919)

## Overview

### BLS Cryptography

BLS signatures allow aggregating multiple signatures into a single signature that can be verified with a single public key, reducing computational overhead in cases where the number of signers is large.

The idea is to aggregate many BLS signatures on a single message (offchain report), as well as using an aggregated public key to verify the message.
The original public keys are not needed for verifying the multi-signature. An important property of the construction is that the scheme is secure against a rogue public-key attack without requiring users to prove knowledge of their secret keys (plain public-key model)

### Network Overview

The example in this package shows how to integrate BLS signatures in a network of networks. Each network is represented by a BLS key, where the nodes in that network are using shares of that key to sign messages, and the public key of that key can be used to verify messages by anyone that is familiar with the public key.

### Key Generation

For simplification, keys are generated in a central way and distributed across nodes in plain text. In a real world scenario, a proper DKG protocol should be used to generate the keys and distribute the shares.

### Signing

Offchain reports are generated in each round by a different node which is the leader of that round, and signed by all nodes in the network. 
Once there is a quorum of valid signatures, the nodes reconstruct a single signature and broadcast to the other networks.

### Verification

Nodes in other networks verify the signature using the public key of the producing network. If the signature is valid, the report is accepted and the round is finalized.

NOTE: nodes also verifies that the report is valid by checking the sequence number against the last finalized round.

## Running the example

```shell
make test-bls
```
