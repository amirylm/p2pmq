package donlib

import (
	"fmt"

	"github.com/smartcontractkit/chainlink/v2/core/services/keystore/chaintype"
	"github.com/smartcontractkit/chainlink/v2/core/services/keystore/keys/ocr2key"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// OcrSigner is a mocked signer that uses `sha256(data)` as signature.
type OcrSigner struct {
	k ocr2key.KeyBundle
}

func NewSigner() *OcrSigner {
	k, err := ocr2key.New(chaintype.EVM)
	if err != nil {
		panic(err)
	}
	return &OcrSigner{
		k: k,
	}
}

func (s *OcrSigner) Sign(reportCtx ocrtypes.ReportContext, report []byte) ([]byte, error) {
	return s.k.Sign(reportCtx, report)
}

func (s *OcrSigner) Verify(pubKey ocrtypes.OnchainPublicKey, reportCtx ocrtypes.ReportContext, report, signed []byte) error {
	if !s.k.Verify(pubKey, reportCtx, report, signed) {
		return fmt.Errorf("invalid report")
	}
	return nil
}
