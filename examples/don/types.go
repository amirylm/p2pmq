package don

import (
	"encoding"
	"fmt"
)

// OracleID is an index over the oracles, used as a succinct attribution to an
// oracle in communication with the on-chain contract. It is not a cryptographic
// commitment to the oracle's private key, like a public key is.
type OracleID uint8

type Report []byte

// OnchainPublicKey is the public key used to cryptographically identify an
// oracle to the on-chain smart contract.
type OnchainPublicKey []byte

// Digest of the configuration for a OCR2 protocol instance. The first two
// bytes indicate which config digester (typically specific to a targeted
// blockchain) was used to compute a ConfigDigest. This value is used as a
// domain separator between different protocol instances and is thus security
// critical. It should be the output of a cryptographic hash function over all
// relevant configuration fields as well as e.g. the address of the target
// contract/state accounts/...
type ConfigDigest [32]byte

func (c ConfigDigest) Hex() string {
	return fmt.Sprintf("%x", c[:])
}

func BytesToConfigDigest(b []byte) (ConfigDigest, error) {
	configDigest := ConfigDigest{}

	if len(b) != len(configDigest) {
		return ConfigDigest{}, fmt.Errorf("cannot convert bytes to ConfigDigest. bytes have wrong length %v", len(b))
	}

	if len(configDigest) != copy(configDigest[:], b) {
		// assertion
		panic("copy returned wrong length")
	}

	return configDigest, nil
}

var _ fmt.Stringer = ConfigDigest{}

func (c ConfigDigest) String() string {
	return c.Hex()
}

var _ encoding.TextMarshaler = ConfigDigest{}

func (c ConfigDigest) MarshalText() (text []byte, err error) {
	s := c.String()
	return []byte(s), nil
}

// ReportTimestamp is the logical timestamp of a report.
type ReportTimestamp struct {
	ConfigDigest ConfigDigest
	Epoch        uint32
	Round        uint8
}

// ReportContext is the contextual data sent to contract along with the report
// itself.
type ReportContext struct {
	ReportTimestamp
	// A hash over some data that is exchanged during execution of the offchain
	// protocol. The data itself is not needed onchain, but we still want to
	// include it in the signature that goes onchain.
	ExtraHash [32]byte
}
