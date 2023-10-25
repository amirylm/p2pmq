package blstest

import (
	"fmt"
	"math/big"

	"github.com/herumi/bls-eth-go-binary/bls"
)

var (
	curveOrder = new(big.Int)
)

// init initializes BLS
func init() {
	_ = bls.Init(bls.BLS12_381)
	_ = bls.SetETHmode(bls.EthModeDraft07)

	curveOrder, _ = curveOrder.SetString(bls.GetCurveOrder(), 10)
}

func GenBlsKey() (*bls.SecretKey, *bls.PublicKey) {
	sk := &bls.SecretKey{}
	sk.SetByCSPRNG()

	return sk, sk.GetPublicKey()
}

// GenShares receives a bls.SecretKey and desired count.
// Will split the secret key into count shares
func GenShares(sk *bls.SecretKey, threshold uint64, count uint64) (map[uint64]*bls.SecretKey, map[uint64]*bls.PublicKey, error) {
	msk := make([]bls.SecretKey, threshold)
	// master key
	msk[0] = *sk

	// construct poly
	for i := uint64(1); i < threshold; i++ {
		sk, _ := GenBlsKey()
		msk[i] = *sk
	}

	// evaluate shares - starting from 1 because 0 is master key
	shares := make(map[uint64]*bls.SecretKey)
	sharesPK := make(map[uint64]*bls.PublicKey, 0)
	for i := uint64(1); i <= count; i++ {
		blsID := bls.ID{}

		err := blsID.SetDecString(fmt.Sprintf("%d", i))
		if err != nil {
			return nil, nil, err
		}

		sk := bls.SecretKey{}
		err = sk.Set(msk, &blsID)
		if err != nil {
			return nil, nil, err
		}
		shares[i] = &sk
		sharesPK[i] = sk.GetPublicKey()
	}

	return shares, sharesPK, nil
}
