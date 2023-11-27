package commons

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaults(t *testing.T) {
	t.Run("Config", func(t *testing.T) {
		cfg := &Config{}
		cfg.Defaults()
		require.Equal(t, 15*time.Second, cfg.DialTimeout)
		require.Equal(t, 1024, cfg.MsgRouterQueueSize)
		require.Equal(t, 10, cfg.MsgRouterWorkers)
		require.Empty(t, cfg.ListenAddrs)
		require.Empty(t, cfg.PSK)
	})

	t.Run("ConnManagerConfig", func(t *testing.T) {
		cfg := &ConnManagerConfig{}
		cfg.Defaults()
		require.Equal(t, 5, cfg.LowWaterMark)
		require.Equal(t, 25, cfg.HighWaterMark)
		require.Equal(t, time.Minute, cfg.GracePeriod)
	})

	t.Run("DiscoveryConfig", func(t *testing.T) {
		cfg := &DiscoveryConfig{}
		cfg.Defaults()
		require.Equal(t, ModeServer, cfg.Mode)
		require.Equal(t, "p2pmq", cfg.ProtocolPrefix)
	})

	t.Run("MsgValidationConfig", func(t *testing.T) {
		cfg := &MsgValidationConfig{}
		cfg.Defaults(nil)
		require.Equal(t, time.Second*5, cfg.Timeout)
		require.Equal(t, 10, cfg.Concurrency)

		cfg.Concurrency = 4
		cfg2 := &MsgValidationConfig{}
		cfg2.Defaults(cfg)
		require.Equal(t, time.Second*5, cfg2.Timeout)
		require.Equal(t, 4, cfg2.Concurrency)
	})
}
