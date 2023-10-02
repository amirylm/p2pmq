# DON simple example

This package shows a high level integration with DONs.

Multiple nodes are created for serving multiple DONs. 
Each node is composed of the following components:
- p2p agent integrated with grpc (`../api/grpc/*`)
- node orchestrator (`./lib/node.go`) to manage the internal processes within the node
- producer (`./lib/transmitter.go`) to publish messages
- message consumer (`./lib/consumer.go`)
- verifier (`./lib/verifier.go`) and signer (`./lib/signer.go`) to verify and sign messages


Within `./tests`, you can find the following tests:

**Local Test**

This test creates a local network of nodes and sends messages between them.
A boostrapper node (kad DHT) is used for peer discovery.

DON orchestrator (`./tests/testdon.go`) is used to manage the DONs and mock consensus rounds that produces dummy reports from a central location, which is helpful for testing.

`Sha256Signer` is used for signing/verifying messages, it doesn't rely on keys and is used for testing with key infrastructure.
