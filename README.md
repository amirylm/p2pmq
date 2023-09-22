# Decentralized Message Engine

<br />

**NOTE: This is an experimental work in progress. DO NOT USE**

<br />

**DME** is a distributed, permissionless messaging engine used for cross oracle communication.

## Overview

A network of agents is capable of the following:
- Broadcast messages over topics with optimal latency
- Pluggable and decoupled message validation using gRPC
- Scoring for protection from bad actors
- Syncing peers with the latest messages to recover from 
restarts, network partition, etc.

![composer-p2pmq.png](./resources/img/composer-p2pmq.png)

<br />
