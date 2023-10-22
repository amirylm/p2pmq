package blstest

import "github.com/herumi/bls-eth-go-binary/bls"

type Share struct {
	SignerID       uint64
	SharePublicKey *bls.PublicKey
	PrivateKey     *bls.SecretKey
	Signers        map[uint64]*bls.PublicKey
}

func (share *Share) Sign(sr *SignedReport) {
	sr.SigHex = share.PrivateKey.SignByte(sr.Data).SerializeToHexStr()
	sr.SignerID = share.SignerID
}

func (share *Share) Validate(sr SignedReport) bool {
	pk, ok := share.Signers[sr.SignerID]
	if !ok {
		return false
	}
	var sign *bls.Sign
	if err := sign.DeserializeHexStr(sr.SigHex); err != nil {
		return false
	}
	if sign.VerifyByte(pk, sr.Data) {
		return true
	}
	return false
}

func (share *Share) QuorumCount() int {
	count := len(share.Signers)
	f := (count - 1) / 3

	return count - f
}
