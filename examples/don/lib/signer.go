package donlib

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

type Signer interface {
	Sign(data []byte) ([]byte, error)
	Verify(signed, pk, data []byte) error
}

// Sha256Signer is a mocked signer that uses `sha256(data)` as signature.
type Sha256Signer struct{}

func (s *Sha256Signer) Sign(data []byte) ([]byte, error) {
	h := sha256.Sum256(data)
	return h[:], nil
}

func (s *Sha256Signer) Verify(signed, _, data []byte) error {
	h := sha256.Sum256(data)
	if !bytes.Equal(signed, h[:]) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}
