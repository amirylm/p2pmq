package blstest

import (
	"context"
	"fmt"

	"github.com/amirylm/p2pmq/api/grpc/client"
	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/proto"
	"github.com/herumi/bls-eth-go-binary/bls"
)

type Node struct {
	utils.StartStopOnce
	threadC utils.ThreadControl

	pubstore *store[*bls.PublicKey]
	reports  *ReportBuffer

	grpc     client.GrpcEndPoint
	verifier client.Verifier
	consumer client.Consumer
	// Shares holds a map of network to the share that is used
	// by this node for that network
	Shares *store[Share]
	// internalReports is a map of network to the reports that are received before quorum is reached
	internalReports *store[*ReportBuffer]
}

func NewNode(grpc client.GrpcEndPoint, pubstore *store[*bls.PublicKey]) *Node {
	n := &Node{
		threadC:         utils.NewThreadControl(),
		reports:         NewReportBuffer(reportBufferSize),
		pubstore:        pubstore,
		grpc:            grpc,
		Shares:          NewStore[Share](),
		internalReports: NewStore[*ReportBuffer](),
	}

	n.verifier = client.NewVerifier(grpc, n.Validate)
	n.consumer = client.NewConsumer(grpc, n.Consume)

	return n
}

func (n *Node) Start() {
	n.StartOnce(func() {
		n.threadC.Go(func(ctx context.Context) {
			err := n.verifier.Start(ctx)
			if err != nil && ctx.Err() == nil {
				panic(err)
			}
		})
		n.threadC.Go(func(ctx context.Context) {
			err := n.consumer.Start(ctx)
			if err != nil && ctx.Err() == nil {
				panic(err)
			}
		})
	})
}

func (n *Node) Close() {
	n.StopOnce(func() {
		n.threadC.Close()
		n.verifier.Stop()
		n.consumer.Stop()
	})
}

func (n *Node) Consume(msg *proto.Message) error {
	sr, err := UnmarshalSignedReport(msg.GetData())
	if err != nil || sr == nil {
		return err
	}
	// if SignerID is GT 0 - it's an internal subnet message
	if sr.SignerID > 0 {
		return n.consumeInternalNet(*sr)
	}
	if !n.reports.Add(sr.Network, *sr) {
		return nil
	}
	fmt.Printf("Added report from network %s\n", sr.Network)

	return nil
}

func (n *Node) Validate(msg *proto.Message) proto.ValidationResult {
	pubkey, ok := n.pubstore.Get(msg.GetTopic())
	if !ok {
		return proto.ValidationResult_IGNORE
	}

	sr, err := UnmarshalSignedReport(msg.GetData())
	if err != nil || sr == nil {
		return proto.ValidationResult_REJECT
	}
	// if SignerID is GT 0 - it's an internal subnet message
	if sr.SignerID > 0 {
		return n.validateInternalNet(*sr)
	}

	sign := &bls.Sign{}
	if err := sign.DeserializeHexStr(sr.SigHex); err != nil {
		return proto.ValidationResult_REJECT
	}

	if !sign.VerifyByte(pubkey, sr.Data) {
		return proto.ValidationResult_REJECT
	}

	return n.reports.ValidateSequence(*sr)
}

func (n *Node) Broadcast(ctx context.Context, report SignedReport) error {
	conn, err := n.grpc.Connect()
	if err != nil {
		return err
	}
	data, err := MarshalSignedReport(&report)
	if err != nil {
		return err
	}
	client := proto.NewControlServiceClient(conn)
	_, err = client.Publish(ctx, &proto.PublishRequest{
		Topic: report.Network,
		Data:  data,
	})
	return err
}

func (n *Node) TriggerReport(report SignedReport) {
	n.threadC.Go(func(ctx context.Context) {
		share, ok := n.Shares.Get(report.Network)
		if !ok {
			return
		}
		if n.getLeader(report.Network, report.SeqNumber) == share.SignerID {
			share.Sign(&report)
			n.reports.Add(report.Network, report)
			if err := n.Broadcast(ctx, report); err != nil {
				fmt.Printf("Error broadcasting report: %s\n", err)
				return
			}
		}
	})
}
