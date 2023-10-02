package donlib

import (
	"context"
	"io"

	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/proto"
)

var (
	skipThreshold    int64 = 5
	invalidThreshold int64 = 25
)

type Verifier interface {
	Start(context.Context) error
	Stop()

	Process(raw []byte) ([]byte, proto.ValidationResult)
}

type verifier struct {
	threadCtrl utils.ThreadControl
	grpc       GrpcEndPoint
	reports    *ReportBuffer
	signer     Signer
}

func NewVerifier(reports *ReportBuffer, grpc GrpcEndPoint, signer Signer) Verifier {
	return &verifier{
		threadCtrl: utils.NewThreadControl(),
		grpc:       grpc,
		reports:    reports,
		signer:     signer,
	}
}

func (v *verifier) Start(ctx context.Context) error {
	conn, err := v.grpc.Connect()
	if err != nil {
		return err
	}
	valRouter := proto.NewValidationRouterClient(conn)
	routerClient, err := valRouter.Handle(ctx)
	if err != nil {
		return err
	}

	valQ := make(chan *proto.Message, 1)

	v.threadCtrl.Go(func(ctx context.Context) {
		defer close(valQ)

		for ctx.Err() == nil {
			msg, err := routerClient.Recv()
			if err == io.EOF || err == context.Canceled || ctx.Err() != nil || msg == nil { // stream closed
				return
			}
			select {
			case <-ctx.Done():
				return
			case valQ <- msg:
			}
		}
	})

	for {
		select {
		case <-ctx.Done():
			return nil
		case next := <-valQ:
			if next == nil {
				return nil
			}
			_, result := v.Process(next.GetData())
			res := &proto.ValidatedMessage{
				Result: result,
				Msg:    next,
			}
			routerClient.Send(res)
		}
	}
}

func (v *verifier) Stop() {
	v.threadCtrl.Close()
}

func (v *verifier) Process(raw []byte) ([]byte, proto.ValidationResult) {
	r, err := UnmarshalReport(raw)
	if err != nil || r == nil {
		// bad encoding
		return raw, proto.ValidationResult_REJECT
	}
	err = v.signer.Verify(r.Sig, nil, r.GetReportData())
	if err != nil {
		return raw, proto.ValidationResult_REJECT
	}

	return raw, v.validateSequence(r)
}

func (v *verifier) validateSequence(r *MockedSignedReport) proto.ValidationResult {
	latest := v.reports.GetLatest(r.Src)
	if latest != nil {
		diff := r.SeqNumber - latest.SeqNumber
		switch {
		case diff > invalidThreshold:
			return proto.ValidationResult_REJECT
		case diff > skipThreshold:
			return proto.ValidationResult_IGNORE
		default: // less than skipThreshold, accept
		}
	}
	if v.reports.Get(r.Src, r.SeqNumber) != nil {
		return proto.ValidationResult_IGNORE
	}
	return proto.ValidationResult_ACCEPT
}
