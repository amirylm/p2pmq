package donlib

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type Signer interface {
	Sign(reportCtx ocrtypes.ReportContext, report []byte) ([]byte, error)
	Verify(pubKey ocrtypes.OnchainPublicKey, reportCtx ocrtypes.ReportContext, report, signed []byte) error
}

// Sha256Signer is a mocked signer that uses `sha256(data)` as signature.
type Sha256Signer struct{}

func (s *Sha256Signer) Sign(_ ocrtypes.ReportContext, report []byte) ([]byte, error) {
	h := sha256.Sum256(report)
	return h[:], nil
}

func (s *Sha256Signer) Verify(_ ocrtypes.OnchainPublicKey, _ ocrtypes.ReportContext, report, signed []byte) error {
	h := sha256.Sum256(report)
	if !bytes.Equal(signed, h[:]) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}
