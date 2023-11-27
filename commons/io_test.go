package commons

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	t.Run("happy path yaml", func(t *testing.T) {
		cfgPath := "resources/config/default.p2pmq.yaml"
		cfg, err := ReadConfig(path.Join("..", cfgPath))
		require.NoError(t, err)
		require.Len(t, cfg.ListenAddrs, 1)
		require.Equal(t, "/ip4/0.0.0.0/tcp/5101", cfg.ListenAddrs[0])
	})

	t.Run("non existing json file", func(t *testing.T) {
		cfgPath := "resources/config/non-existing.p2pmq.json"
		_, err := ReadConfig(path.Join("..", cfgPath))
		require.Error(t, err)
	})
}
