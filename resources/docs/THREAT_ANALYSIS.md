# Threat Analysis

| Title | Description | Severity | Impact | Mitigation |
| ----- | ----------- | -------- | ------ | ---------- |
| Message Spamming | Attackers flood the network with invalid or malicious messages, consuming bandwidth and degrading network performance. | Very high | Network congestion, performance degradation, resource exhaustion. | - Require message validation with cryptographic signatures to ensure message authenticity and integrity \ - Require message sequence validation, where unrealistic sequences are considered invalid. Use caching, where the key is based on content hashing, so the same message won't exhaust resources. |
| Message Spamming: Validation Queue Flooding | Attackers can overload the validation queue by sending spam messages at a very high rate. Legitimate messages get dropped, resulting in a denial of service as messages are ignored. | Very high | Denial of service, message loss. | Implement a circuit breaker before the validation queue that makes informed decisions based on message origin IP and a probabilistic strategy to drop messages. See [gossipsub v1.1: validation-queue-protection](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/red.md#gossipsub-v11-functional-extension-for-validation-queue-protection).
| Censorship Attacks | Malicious nodes selectively block messages to suppress certain information. Countermeasures include redundancy and diversification of message propagation paths. | Very high | Information suppression, network manipulation. | - Maintain a diverse set of mesh peers to maintain network resilience \ - Use redundancy in message propagation paths to counter censorship attacks \ - Employ mechanisms to detect and mitigate Sybil nodes, such as peer scoring and validation |
| Denial of Service (DoS) | Adversaries flood the network with malicious traffic or connections to disrupt its operation. | High | Network disruption, resource exhaustion. | Implement rate limiting, connection policies, and adaptive firewall mechanisms to protect against DoS attacks. |
| Partition Attacks | Adversaries attempt to partition the network by disrupting communication between mesh peers. | High | Network fragmentation, reduced communication. | Ensuring a diverse set of well-behaved mesh peers can help prevent this. Implement a robust peer scoring system to detect and disconnect poorly performing or malicious peers. See [gossipsub v1.1: peer scoring](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.1.md#peer-scoring).
| Sybil Attacks | Attackers create multiple fake identities (Sybil nodes) to manipulate the network. | Medium | Potential network manipulation, disruption. | Relies on the effectiveness of peer discovery mechanisms, including whitelisting which in controlled by the parent node, which should have access to that information. |
| Eclipse Attacks | Malicious nodes attempt to control a target node's connections, isolating it from the legitimate peers. | Medium | Network isolation, potential data manipulation. | Ensure diverse connectivity by utilizing peer discovery methods and continuously change connected peers. |
| DHT Pollution: connections | Malicious nodes flood the DHT with malicious entries as part of an eclipse attacke | Medium | Network isolation. | - Implement DHT security mechanisms to prevent unauthorized writes and ensure data validity \ - Regularly check the integrity of DHT data and remove or quarantine polluted entries |
| DHT Pollution: storage | Malicious nodes flood the DHT with irrelevant data, potentially disrupting the network's ability to perform efficient content retrieval. | Medium | Degraded performance in content retrieval, network congestion, resource exhaustion. | - Implement DHT security mechanisms to prevent unauthorized writes and ensure data validity \ - Regularly check the integrity of DHT data and remove or quarantine polluted entries \ - Implement rate limiting and access controls for DHT writes to mitigate pollution attempts. |


### Message Spamming

Attackers flood the network with invalid or malicious messages, consuming bandwidth and degrading network performance.

**Severity:** Very high 
**Impact:** Network congestion, performance degradation, resource exhaustion.

**Mitigation strategy:**

- Require message validation with cryptographic signatures to ensure message authenticity and integrity
- Require message sequence validation, where unrealistic sequences are considered invalid
- Use caching, where the key is based on content hashing, so the same message won't exhaust resources.

### Message Spamming: Validation Queue Flooding

Attackers can overload the validation queue by sending spam messages at a very high rate. Legitimate messages get dropped, resulting in a denial of service as messages are ignored.

**Severity:** Very high
**Impact:** Denial of service, message loss.

**Mitigation strategy:**
Implement a circuit breaker before the validation queue that makes informed decisions based on message origin IP and a probabilistic strategy to drop messages. See [gossipsub v1.1: validation-queue-protection](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/red.md#gossipsub-v11-functional-extension-for-validation-queue-protection).

### Censorship Attacks

Malicious nodes selectively block messages to suppress certain information. Countermeasures include redundancy and diversification of message propagation paths.

**Severity:** Very high
**Impact:** Information suppression, network manipulation.

**Mitigation Strategies:**

- Maintain a diverse set of mesh peers to maintain network resilience
- Use redundancy in message propagation paths to counter censorship attacks
- Employ mechanisms to detect and mitigate Sybil nodes, such as peer scoring and validation

### Denial of Service (DoS)

Adversaries flood the network with malicious traffic or connections to disrupt its operation.

**Severity:** High
**Impact:** Network disruption, resource exhaustion.

**Mitigation strategy:**
Implement rate limiting, connection policies, and adaptive firewall mechanisms to protect against DoS attacks.

### Partition Attacks

Adversaries attempt to partition the network by disrupting communication between mesh peers.

**Severity:** High
**Impact:** Network fragmentation, reduced communication.

**Mitigation Strategy:**
Ensuring a diverse set of well-behaved mesh peers can help prevent this.
Implement a robust peer scoring system to detect and disconnect poorly performing or malicious peers. See [gossipsub v1.1: peer scoring](https://github.com/libp2p/specs/blob/master/pubsub/gossipsub/gossipsub-v1.1.md#peer-scoring).

### Sybil Attacks

Attackers create multiple fake identities (Sybil nodes) to manipulate the network.

**Severity:** Medium
**Impact:** Potential network manipulation, disruption.

**Mitigation strategy:**
Relies on the effectiveness of peer discovery mechanisms, including whitelisting which in controlled by the parent node, which should have access to that information.

### Eclipse Attacks

Malicious nodes attempt to control a target node's connections, isolating it from the legitimate peers.

**Severity:** Medium
**Impact:** Network isolation, potential data manipulation.

**Mitigation strategy:**
Ensure diverse connectivity by utilizing peer discovery methods and continuously change connected peers.

### DHT Pollution

Malicious nodes flood the DHT with malicious entries as part of an eclipse attack or irrelevant data, potentially disrupting the network's ability to perform efficient content retrieval (NOTE: in case the DHT will be used for storage in addition to discovery).

**Severity:** Medium
**Impact:** Network isolation, degraded performance in content retrieval, network congestion, resource exhaustion.

**Mitigation Strategies:**

- Implement DHT security mechanisms to prevent unauthorized writes and ensure data validity
- Regularly check the integrity of DHT data and remove or quarantine polluted entries
- Implement rate limiting and access controls for DHT writes to mitigate pollution attempts.