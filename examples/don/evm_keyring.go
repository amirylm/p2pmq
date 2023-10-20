package don

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var curve = secp256k1.S256()

func RawReportContext(repctx ReportContext) [3][32]byte {
	rawRepctx := [3][32]byte{}
	copy(rawRepctx[0][:], repctx.ConfigDigest[:])
	binary.BigEndian.PutUint32(rawRepctx[1][32-5:32-1], repctx.Epoch)
	rawRepctx[1][31] = repctx.Round
	rawRepctx[2] = repctx.ExtraHash
	return rawRepctx
}

type EvmKeyring struct {
	privateKey ecdsa.PrivateKey
}

func NewEVMKeyring(material io.Reader) (*EvmKeyring, error) {
	ecdsaKey, err := ecdsa.GenerateKey(curve, material)
	if err != nil {
		return nil, err
	}
	return &EvmKeyring{privateKey: *ecdsaKey}, nil
}

// XXX: PublicKey returns the address of the public key not the public key itself
func (ok *EvmKeyring) PublicKey() OnchainPublicKey {
	address := ok.signingAddress()
	return address[:]
}

func (ok *EvmKeyring) OnChainPublicKey() string {
	return hex.EncodeToString(ok.PublicKey())
}

func (ok *EvmKeyring) reportToSigData(reportCtx ReportContext, report Report) []byte {
	rawReportContext := RawReportContext(reportCtx)
	sigData := crypto.Keccak256(report)
	sigData = append(sigData, rawReportContext[0][:]...)
	sigData = append(sigData, rawReportContext[1][:]...)
	sigData = append(sigData, rawReportContext[2][:]...)
	return crypto.Keccak256(sigData)
}

func (ok *EvmKeyring) Sign(reportCtx ReportContext, report Report) ([]byte, error) {
	return crypto.Sign(ok.reportToSigData(reportCtx, report), &ok.privateKey)
}

func (ok *EvmKeyring) Verify(publicKey OnchainPublicKey, reportCtx ReportContext, report Report, signature []byte) bool {
	hash := ok.reportToSigData(reportCtx, report)
	authorPubkey, err := crypto.SigToPub(hash, signature)
	if err != nil {
		return false
	}
	authorAddress := crypto.PubkeyToAddress(*authorPubkey)
	return bytes.Equal(publicKey[:], authorAddress[:])
}

func (ok *EvmKeyring) MaxSignatureLength() int {
	return 65
}

func (ok *EvmKeyring) signingAddress() common.Address {
	return crypto.PubkeyToAddress(*(&ok.privateKey).Public().(*ecdsa.PublicKey))
}

func (ok *EvmKeyring) Marshal() ([]byte, error) {
	return crypto.FromECDSA(&ok.privateKey), nil
}

func (ok *EvmKeyring) Unmarshal(in []byte) error {
	privateKey, err := crypto.ToECDSA(in)
	if err != nil {
		return err
	}
	ok.privateKey = *privateKey
	return nil
}
