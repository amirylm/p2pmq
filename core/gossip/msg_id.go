package gossip

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
)

type MsgIDSize int
type MsgIDFuncType string

const (
	MsgIDSha256Type MsgIDFuncType = "sha256"
	MsgIDMD5Type    MsgIDFuncType = "md5"
)

var DefaultMsgIDFn = MsgIDSha256(20)

func MsgIDFn(tp MsgIDFuncType, size MsgIDSize) pubsub.MsgIdFunction {
	switch tp {
	case MsgIDSha256Type:
		return MsgIDSha256(size)
	case MsgIDMD5Type:
		return MsgIDMD5(size)
	default:
		return DefaultMsgIDFn
	}
}

// MsgIDSha256 uses sha256 hash of the message content
func MsgIDSha256(size MsgIDSize) pubsub.MsgIdFunction {
	return func(pmsg *pubsub_pb.Message) string {
		msg := pmsg.GetData()
		if len(msg) == 0 {
			return ""
		}
		// TODO: optimize, e.g. by using a pool of hashers
		h := sha256.Sum256(msg)
		if msgSize := MsgIDSize(len(h)); size > msgSize {
			size = msgSize
		}
		return hex.EncodeToString(h[:size])
	}
}

// MsgIDSMD5 uses md5 hash of the message content
func MsgIDMD5(size MsgIDSize) pubsub.MsgIdFunction {
	return func(pmsg *pubsub_pb.Message) string {
		msg := pmsg.GetData()
		if len(msg) == 0 {
			return ""
		}
		// TODO: optimize, e.g. by using a pool of hashers
		h := md5.Sum(msg)
		if msgSize := MsgIDSize(len(h)); size > msgSize {
			size = msgSize
		}
		return hex.EncodeToString(h[:size])
	}
}
