package gossip

import (
	"testing"

	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/stretchr/testify/require"
)

func TestMsgID(t *testing.T) {
	tests := []struct {
		name  string
		msgID MsgIDFuncType
		size  MsgIDSize
		input []byte
		want  string
	}{
		{
			name:  "sha256",
			msgID: MsgIDSha256Type,
			size:  20,
			input: []byte("hello world"),
			want:  "b94d27b9934d3e08a52e52d7da7dabfac484efe3",
		},
		{
			name:  "md5",
			msgID: MsgIDMD5Type,
			size:  10,
			input: []byte("hello world"),
			want:  "5eb63bbbe01eeed093cb",
		},
		{
			name:  "default",
			msgID: "",
			size:  0,
			input: []byte("hello world"),
			want:  "b94d27b9934d3e08a52e52d7da7dabfac484efe3",
		},
		{
			name:  "size overflow",
			msgID: MsgIDMD5Type,
			size:  100,
			input: []byte("hello world"),
			want:  "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgIDFn := MsgIDFn(tc.msgID, tc.size)
			got := msgIDFn(&pubsub_pb.Message{
				Data: tc.input,
			})
			require.Equal(t, tc.want, got)
		})
	}
}
