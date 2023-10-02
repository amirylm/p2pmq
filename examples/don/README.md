# DON simple example

This package shows a high level integration with DONs.

Multiple nodes are created for serving multiple DONs. 
Each node is composed of the following components:
- p2p agent integrated with grpc (`../api/grpc/*`)
- orchestrator (`./lib/node.go`) to manage the internal processes within the node
- producer (`./lib/transmitter.go`) to publish messages
- message consumer (`./lib/consumer.go`)
- verifier (`./lib/verifier.go`) and signer (`./lib/signer.go`) to verify and sign messages

The nodes are connected to each other via a local p2p network.
A boostrapper node (kad DHT) is used for peer discovery, providing an entry point to the network.