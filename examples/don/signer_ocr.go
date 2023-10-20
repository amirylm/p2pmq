package don

import (
	cryptorand "crypto/rand"
	"fmt"
)

// OcrSigner is a mocked signer that uses `sha256(data)` as signature.
type OcrSigner struct {
	k   *EvmKeyring
	oid OracleID
}

func NewSigner(oid OracleID) *OcrSigner {
	k, err := NewEVMKeyring(cryptorand.Reader)
	if err != nil {
		panic(err)
	}
	return &OcrSigner{
		k:   k,
		oid: oid,
	}
}

func (s *OcrSigner) OracleID() OracleID {
	return s.oid
}

func (s *OcrSigner) Sign(reportCtx ReportContext, report []byte) ([]byte, error) {
	return s.k.Sign(reportCtx, report)
}

func (s *OcrSigner) Verify(pubKey OnchainPublicKey, reportCtx ReportContext, report, signed []byte) error {
	if !s.k.Verify(pubKey, reportCtx, report, signed) {
		return fmt.Errorf("invalid report")
	}
	return nil
}
