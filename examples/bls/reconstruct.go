package blstest

import (
	"fmt"

	"github.com/herumi/bls-eth-go-binary/bls"
)

// ReconstructSigs receives a map of operator indexes and the corresponding
// serialized bls.Sign. It then reconstructs the original threshold signature
// using lagrange interpolation.
func ReconstructSigs(signatures map[uint64][]byte) (*bls.Sign, error) {
	reconstructed := bls.Sign{}

	idVec := make([]bls.ID, 0)
	sigVec := make([]bls.Sign, 0)

	for index, signature := range signatures {
		id := bls.ID{}
		err := id.SetDecString(fmt.Sprintf("%d", index))
		if err != nil {
			return nil, err
		}

		idVec = append(idVec, id)
		sig := bls.Sign{}

		err = sig.Deserialize(signature)
		if err != nil {
			return nil, err
		}

		sigVec = append(sigVec, sig)
	}
	err := reconstructed.Recover(sigVec, idVec)

	return &reconstructed, err
}
