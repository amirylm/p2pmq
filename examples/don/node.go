package don

import (
	"context"

	"github.com/amirylm/p2pmq/api/grpc/client"
	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/proto"
)

type Node struct {
	utils.StartStopOnce
	threadC utils.ThreadControl

	Grpc    client.GrpcEndPoint
	Reports *ReportBuffer

	Transmitter OverlayTrasnmitter
	// Consumer is the consumer client to get messages from the MsgRouter
	Consumer client.Consumer
	// Verifier is the verifier client that faciliates interaction with the ValidationRouter
	Verifier client.Verifier
	// reportVerifier is doing the actual validation on incoming messages
	reportVerifier *reportVerifier
	Signers        map[string]Signer
}

func NewNode(grpc client.GrpcEndPoint, opts ...NodeOpt) *Node {
	nodeOpts := defaultNodeOpts()
	for _, opt := range opts {
		opt(nodeOpts)
	}
	reports := NewReportBuffer(nodeOpts.bufferSize)
	reportVerifier := NewReportVerifier(reports)
	return &Node{
		threadC: utils.NewThreadControl(),

		Grpc:    grpc,
		Reports: reports,

		Transmitter: NewOverlayTransmitter(grpc, nodeOpts.reportManipulator),
		Consumer: client.NewConsumer(grpc, func(msg *proto.Message) error {
			r, err := UnmarshalReport(msg.GetData())
			if err != nil || r == nil {
				// bad encoding, not expected
				return err
			}
			if reports.Add(r.Src, *r) {
				// TODO: notify new report
			}
			return nil
		}),
		Verifier: client.NewVerifier(grpc, func(msg *proto.Message) proto.ValidationResult {
			_, res := reportVerifier.Process(msg)
			return res
		}),
		Signers: map[string]Signer{},
	}
}

func (n *Node) Start() {
	n.StartOnce(func() {
		n.threadC.Go(func(ctx context.Context) {
			defer n.Consumer.Stop()
			if err := n.Consumer.Start(ctx); err != nil {
				panic(err)
			}
		})
		n.threadC.Go(func(ctx context.Context) {
			defer n.Verifier.Stop()
			if err := n.Verifier.Start(ctx); err != nil {
				panic(err)
			}
		})
	})
}

func (n *Node) Stop() {
	n.StopOnce(func() {
		n.threadC.Close()
	})
}

type nodeOpts struct {
	reportManipulator func(*MockedSignedReport)
	bufferSize        int
	signer            Signer
}

func defaultNodeOpts() *nodeOpts {
	return &nodeOpts{
		reportManipulator: func(*MockedSignedReport) {},
		bufferSize:        1024,
	}
}

type NodeOpt func(*nodeOpts)

func WithReportManipulator(reportManipulator func(*MockedSignedReport)) NodeOpt {
	return func(opts *nodeOpts) {
		opts.reportManipulator = reportManipulator
	}
}

func WithBufferSize(bufferSize int) NodeOpt {
	return func(opts *nodeOpts) {
		opts.bufferSize = bufferSize
	}
}

func WithSigner(signer Signer) NodeOpt {
	return func(opts *nodeOpts) {
		opts.signer = signer
	}
}
