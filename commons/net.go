package commons

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"
)

func GetOrGeneratePrivateKey(privKeyB64 string) (sk crypto.PrivKey, encodedB64 string, err error) {
	if len(privKeyB64) == 0 {
		sk, _, err = crypto.GenerateEd25519Key(crand.Reader)
		if err != nil {
			return nil, encodedB64, fmt.Errorf("could not generate private key: %w", err)
		}
		encoded, err := crypto.MarshalPrivateKey(sk)
		if err != nil {
			return nil, encodedB64, err
		}
		// TODO: pool base64 encoders
		encodedB64 = base64.StdEncoding.EncodeToString(encoded)
		return sk, encodedB64, nil
	}
	encoded, err := base64.StdEncoding.DecodeString(encodedB64)
	if err != nil {
		return nil, privKeyB64, fmt.Errorf("failed to decode private key with base64: %w", err)
	}
	sk, err = crypto.UnmarshalPrivateKey(encoded)
	if err != nil {
		return nil, privKeyB64, fmt.Errorf("failed to unmarshal private key: %w", err)
	}
	return sk, encodedB64, nil
}
