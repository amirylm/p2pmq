package donlib

import (
	"context"

	"github.com/amirylm/p2pmq/commons/utils"
)

type Node struct {
	utils.StartStopOnce
	threadC utils.ThreadControl

	Grpc    GrpcEndPoint
	Reports *ReportBuffer

	Transmitter OverlayTrasnmitter
	Consumer    MsgConsumer
	Verifier    Verifier
	Signer      Signer
}

func NewNode(grpc GrpcEndPoint, opts ...NodeOpt) *Node {
	nodeOpts := defaultNodeOpts()
	for _, opt := range opts {
		opt(nodeOpts)
	}
	reports := NewReportBuffer(nodeOpts.bufferSize)
	return &Node{
		threadC: utils.NewThreadControl(),

		Grpc:    grpc,
		Reports: reports,

		Transmitter: NewOverlayTransmitter(grpc, nodeOpts.reportManipulator),
		Consumer:    NewMsgConsumer(reports, grpc),
		Verifier:    NewVerifier(reports, grpc, nodeOpts.signer),
		Signer:      nodeOpts.signer,
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
		signer:            &Sha256Signer{},
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
