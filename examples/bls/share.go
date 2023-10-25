package blstest

import (
	"fmt"

	"github.com/herumi/bls-eth-go-binary/bls"
)

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
	fmt.Printf("Signer %d is validating report on network %s with seq %d from %d\n", share.SignerID, sr.Network, sr.SeqNumber, sr.SignerID)
	sign := &bls.Sign{}
	if err := sign.DeserializeHexStr(sr.SigHex); err != nil {
		return false
	}
	if sign.VerifyByte(pk, sr.Data) {
		return true
	}
	return false
}

func (share *Share) QuorumCount() int {
	return Threshold(len(share.Signers))
}

func Threshold(count int) int {
	f := (count - 1) / 3

	return count - f
}
