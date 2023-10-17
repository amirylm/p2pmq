module github.com/amirylm/p2pmq/examples/don

go 1.20

require github.com/smartcontractkit/chainlink/v2 v2.5.0

require (
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/amirylm/p2pmq v0.0.0-20231003142105-54833f1b93eb // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd v0.23.4 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deckarep/golang-set/v2 v2.3.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/ethereum/go-ethereum v1.12.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gogo/protobuf v1.3.3 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hdevalence/ed25519consensus v0.1.0 // indirect
	github.com/holiman/uint256 v1.2.2 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-libp2p-core v0.8.5 // indirect
	github.com/libp2p/go-openssl v0.0.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.11.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/smartcontractkit/caigo v0.0.0-20230621050857-b29a4ca8c704 // indirect
	github.com/smartcontractkit/libocr v0.0.0-20230816220705-665e93233ae5 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/tidwall/gjson v1.14.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.12.0 // indirect
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63 // indirect
	golang.org/x/net v0.14.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.12.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
	google.golang.org/grpc v1.56.2 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	// Fix go mod tidy issue for ambiguous imports from go-ethereum
	// See https://github.com/ugorji/go/issues/279
	github.com/btcsuite/btcd => github.com/btcsuite/btcd v0.22.1

	// replicating the replace directive on cosmos SDK
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

	// until merged upstream: https://github.com/hashicorp/go-plugin/pull/257
	github.com/hashicorp/go-plugin => github.com/smartcontractkit/go-plugin v0.0.0-20230605132010-0f4d515d1472

	// until merged upstream: https://github.com/mwitkow/grpc-proxy/pull/69
	github.com/mwitkow/grpc-proxy => github.com/smartcontractkit/grpc-proxy v0.0.0-20230731113816-f1be6620749f
)
