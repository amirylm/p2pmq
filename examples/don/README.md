# DON simple example

This package shows a high level integration with DONs.

Multiple nodes are created for serving multiple DONs. 
Each node is composed of the following components:
- p2p agent integrated with grpc (`../../api/grpc/*`)
- node orchestrator (`./node.go`) to manage the internal processes within the node
- don orchestrator (`./localdon.go`) to facilitate the behaviour of a single don, locally within the same process. 
- producer (`./transmitter.go`) to publish messages
- message consumer (`./consumer.go`)
- verifier (`./verifier.go`) and signer (`./signer.go`) to verify and sign messages


You can find the following tests:

**Local Test**

This test creates a local network of nodes and sends messages between them.
A boostrapper node (kad DHT) is used for peer discovery.

DON orchestrator (`./localdon.go`) is used to manage the DONs and mock consensus rounds that produces dummy reports from a central location, which is helpful for testing.


## Running the example

```shell
make test-localdon
```
