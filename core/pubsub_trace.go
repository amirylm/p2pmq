package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsub_pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

type traceFaucets struct {
	join, leave, publish, deliver, reject, duplicate, addPeer, removePeer, graft, prune, sendRPC, dropRPC, recvRPC *atomic.Uint64
}

func newTraceFaucets() traceFaucets {
	return traceFaucets{
		join:       new(atomic.Uint64),
		leave:      new(atomic.Uint64),
		publish:    new(atomic.Uint64),
		deliver:    new(atomic.Uint64),
		reject:     new(atomic.Uint64),
		duplicate:  new(atomic.Uint64),
		addPeer:    new(atomic.Uint64),
		removePeer: new(atomic.Uint64),
		graft:      new(atomic.Uint64),
		prune:      new(atomic.Uint64),
		sendRPC:    new(atomic.Uint64),
		dropRPC:    new(atomic.Uint64),
		recvRPC:    new(atomic.Uint64),
	}
}

func (tf *traceFaucets) add(other traceFaucets) {
	tf.join.Add(other.join.Load())
	tf.leave.Add(other.leave.Load())
	tf.publish.Add(other.publish.Load())
	tf.deliver.Add(other.deliver.Load())
	tf.reject.Add(other.reject.Load())
	tf.duplicate.Add(other.duplicate.Load())
	tf.addPeer.Add(other.addPeer.Load())
	tf.removePeer.Add(other.removePeer.Load())
	tf.graft.Add(other.graft.Load())
	tf.prune.Add(other.prune.Load())
	tf.sendRPC.Add(other.sendRPC.Load())
	tf.dropRPC.Add(other.dropRPC.Load())
	tf.recvRPC.Add(other.recvRPC.Load())
}

type eventFields map[string]string

func MarshalTraceEvents(events []eventFields) ([]byte, error) {
	return json.Marshal(events)
}

func UnmarshalTraceEvents(data []byte) ([]eventFields, error) {
	var events []eventFields
	err := json.Unmarshal(data, &events)
	return events, err
}

// psTracer helps to trace pubsub events, implements pubsublibp2p.EventTracer
type psTracer struct {
	lggr      *zap.SugaredLogger
	subTracer pubsub.EventTracer
	skiplist  []string
	faucets   traceFaucets
	lock      sync.Mutex
	events    []eventFields
	debug     bool
}

// NewTracer creates an instance of pubsub tracer
func newPubsubTracer(lggr *zap.SugaredLogger, debug bool, skiplist []string, subTracer pubsub.EventTracer) pubsub.EventTracer {
	return &psTracer{
		lggr:      lggr.Named("PubsubTracer"),
		subTracer: subTracer,
		skiplist:  skiplist,
		debug:     debug,
		faucets:   newTraceFaucets(),
	}
}

func (pst *psTracer) Reset() {
	pst.lock.Lock()
	defer pst.lock.Unlock()

	pst.events = nil
	pst.faucets = traceFaucets{}
}

func (pst *psTracer) Events() []eventFields {
	pst.lock.Lock()
	defer pst.lock.Unlock()

	return pst.events
}

func (pst *psTracer) Faucets() traceFaucets {
	return pst.faucets
}

// Trace handles events, implementation of pubsub.EventTracer
func (pst *psTracer) Trace(evt *pubsub_pb.TraceEvent) {
	fields := eventFields{}
	fields["type"] = evt.GetType().String()
	fields["time"] = time.Now().Format(time.RFC3339)
	pid, err := peer.IDFromBytes(evt.GetPeerID())
	if err != nil {
		fields["peerID"] = "error"
	}
	fields["peerID"] = pid.String()
	eventType := evt.GetType()
	switch eventType {
	case pubsub_pb.TraceEvent_PUBLISH_MESSAGE:
		pst.faucets.publish.Add(1)
		msg := evt.GetPublishMessage()
		evt.GetPeerID()
		fields["msgID"] = hex.EncodeToString(msg.GetMessageID())
		fields["topic"] = msg.GetTopic()
	case pubsub_pb.TraceEvent_REJECT_MESSAGE:
		pst.faucets.reject.Add(1)
		msg := evt.GetRejectMessage()
		pid, err := peer.IDFromBytes(msg.GetReceivedFrom())
		if err == nil {
			fields["receivedFrom"] = pid.String()
		}
		fields["msgID"] = hex.EncodeToString(msg.GetMessageID())
		fields["topic"] = msg.GetTopic()
		fields["reason"] = msg.GetReason()
	case pubsub_pb.TraceEvent_DUPLICATE_MESSAGE:
		pst.faucets.duplicate.Add(1)
		msg := evt.GetDuplicateMessage()
		pid, err := peer.IDFromBytes(msg.GetReceivedFrom())
		if err == nil {
			fields["receivedFrom"] = pid.String()
		}
		fields["msgID"] = hex.EncodeToString(msg.GetMessageID())
		fields["topic"] = msg.GetTopic()
	case pubsub_pb.TraceEvent_DELIVER_MESSAGE:
		pst.faucets.deliver.Add(1)
		msg := evt.GetDeliverMessage()
		pid, err := peer.IDFromBytes(msg.GetReceivedFrom())
		if err == nil {
			fields["receivedFrom"] = pid.String()
		}
		fields["msgID"] = hex.EncodeToString(msg.GetMessageID())
		fields["topic"] = msg.GetTopic()
	case pubsub_pb.TraceEvent_ADD_PEER:
		pst.faucets.addPeer.Add(1)
		pid, err := peer.IDFromBytes(evt.GetAddPeer().GetPeerID())
		if err == nil {
			fields["targetPeer"] = pid.String()
		}
	case pubsub_pb.TraceEvent_REMOVE_PEER:
		pst.faucets.removePeer.Add(1)
		pid, err := peer.IDFromBytes(evt.GetRemovePeer().GetPeerID())
		if err == nil {
			fields["targetPeer"] = pid.String()
		}
	case pubsub_pb.TraceEvent_JOIN:
		pst.faucets.join.Add(1)
		fields["topic"] = evt.GetJoin().GetTopic()
	case pubsub_pb.TraceEvent_LEAVE:
		pst.faucets.leave.Add(1)
		fields["topic"] = evt.GetLeave().GetTopic()
	case pubsub_pb.TraceEvent_GRAFT:
		pst.faucets.graft.Add(1)
		msg := evt.GetGraft()
		pid, err := peer.IDFromBytes(msg.GetPeerID())
		if err == nil {
			fields["graftPeer"] = pid.String()
		}
		fields["topic"] = msg.GetTopic()
	case pubsub_pb.TraceEvent_PRUNE:
		pst.faucets.prune.Add(1)
		msg := evt.GetPrune()
		pid, err := peer.IDFromBytes(msg.GetPeerID())
		if err == nil {
			fields["prunePeer"] = pid.String()
		}
		fields["topic"] = msg.GetTopic()
	case pubsub_pb.TraceEvent_SEND_RPC:
		pst.faucets.sendRPC.Add(1)
		msg := evt.GetSendRPC()
		pid, err := peer.IDFromBytes(msg.GetSendTo())
		if err == nil {
			fields["targetPeer"] = pid.String()
		}
		if meta := msg.GetMeta(); meta != nil {
			if ctrl := meta.Control; ctrl != nil {
				fields = appendIHave(fields, ctrl.GetIhave())
				fields = appendIWant(fields, "self", ctrl.GetIwant())
				// ctrl.GetGraft()
				// ctrl.GetPrune()
			}
			var subs []string
			for _, sub := range meta.Subscription {
				subs = append(subs, sub.GetTopic())
			}
			fields["subs"] = strings.Join(subs, ",")
		}
	case pubsub_pb.TraceEvent_DROP_RPC:
		pst.faucets.dropRPC.Add(1)
		msg := evt.GetDropRPC()
		pid, err := peer.IDFromBytes(msg.GetSendTo())
		if err == nil {
			fields["targetPeer"] = pid.String()
		}
	case pubsub_pb.TraceEvent_RECV_RPC:
		pst.faucets.recvRPC.Add(1)
		msg := evt.GetRecvRPC()
		pid, err := peer.IDFromBytes(msg.GetReceivedFrom())
		if err == nil {
			fields["receivedFrom"] = pid.String()
		}
		if meta := msg.GetMeta(); meta != nil {
			if ctrl := meta.Control; ctrl != nil {
				fields = appendIHave(fields, ctrl.GetIhave())
				fields = appendIWant(fields, pid.String(), ctrl.GetIwant())
			}
			var subs []string
			for _, sub := range meta.Subscription {
				subs = append(subs, sub.GetTopic())
			}
			fields["subs"] = strings.Join(subs, ",")
		}
	default:
	}

	if pst.shouldSkip(eventType.String()) {
		return
	}

	pst.debugEvent(fields)
	pst.storeEvent(fields)

	if pst.subTracer != nil {
		pst.subTracer.Trace(evt)
	}
}

func (pst *psTracer) debugEvent(fields eventFields) {
	if pst.debug {
		pst.lggr.Debugf("pubsub trace event: %+v", fields)
	}
}

func (pst *psTracer) storeEvent(fields eventFields) {
	pst.lock.Lock()
	defer pst.lock.Unlock()

	pst.events = append(pst.events, fields)
}

func (pst *psTracer) shouldSkip(eventType string) bool {
	pst.lock.Lock()
	defer pst.lock.Unlock()

	for _, skip := range pst.skiplist {
		if eventType == skip {
			return true
		}
	}
	return false
}

func appendIHave(fields map[string]string, ihave []*pubsub_pb.TraceEvent_ControlIHaveMeta) map[string]string {
	if len(ihave) > 0 {
		fields["ihaveCount"] = strconv.Itoa(len(ihave))
		for _, im := range ihave {
			var mids []string
			msgids := im.GetMessageIDs()
			for _, mid := range msgids {
				mids = append(mids, hex.EncodeToString(mid))
			}
			fields[fmt.Sprintf("%s-IHAVEmsgIDs", im.GetTopic())] = strings.Join(mids, ",")
		}
	}
	return fields
}

func appendIWant(fields map[string]string, peer string, iwant []*pubsub_pb.TraceEvent_ControlIWantMeta) map[string]string {
	if len(iwant) > 0 {
		fields["iwantCount"] = strconv.Itoa(len(iwant))
		var mids []string
		for _, im := range iwant {
			msgids := im.GetMessageIDs()
			for _, mid := range msgids {
				mids = append(mids, hex.EncodeToString(mid))
			}
		}
		fields[fmt.Sprintf("%s-IWANTmsgIDs", peer)] = strings.Join(mids, ",")
	}
	return fields
}
