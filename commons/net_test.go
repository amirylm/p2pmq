package commons

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetOrGeneratePrivateKey(t *testing.T) {
	sk, skb64, err := GetOrGeneratePrivateKey("")
	require.NoError(t, err)
	require.NotNil(t, sk)
	require.NotEmpty(t, skb64)

	sk2, sk2b64, err := GetOrGeneratePrivateKey(skb64)
	require.NoError(t, err)
	require.NotNil(t, sk2)
	require.Equal(t, skb64, sk2b64)
	require.True(t, sk.Equals(sk2))
}
