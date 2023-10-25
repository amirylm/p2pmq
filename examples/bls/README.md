# BLS example

This package shows a high level integration with BLS based networks.

## Overview

Every internal network is represented by a BLS key, which is splitted into shares and distributed among the nodes in that network.
Internal messages are signed with the shares of nodes, once a threshold of signatures is reached, the nodes of that network reconstruct the signature and broadcast it to the other networks. Nodes in other networks verifies the messages using the public key of the BLS key of the producing network.

## Running the example

```shell
make test-bls
```