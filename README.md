# Decentralized Message Engine

<br />

**NOTE: This is an experimental work in progress. DO NOT USE**

## Overview

**DME** is a distributed, permissionless messaging engine for cross oracle communication.

A network of agents is capable of the following:
- Broadcast messages over topics with optimal latency
- Pluggable and decoupled message validation using gRPC
- Scoring for protection from bad actors
- Syncing peers with the latest messages to recover from 
restarts, network partition, etc.
