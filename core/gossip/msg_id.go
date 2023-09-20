package gossip

import (
	"crypto/sha256"
	"encoding/hex"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
)

// msgIDSha256 uses sha256 hash of the message content
func MsgIDSha256(size int) pubsub.MsgIdFunction {
	return func(pmsg *pubsub_pb.Message) string {
		msg := pmsg.GetData()
		if len(msg) == 0 {
			return ""
		}
		// TODO: optimize, e.g. by using a pool of hashers
		h := sha256.Sum256(msg)
		return hex.EncodeToString(h[:size])
	}
}
