package don

import (
	"context"
	"fmt"

	"github.com/amirylm/p2pmq/proto"
)

func donTopic(donID string) string {
	return fmt.Sprintf("don.%s", donID)
}

type OverlayTrasnmitter interface {
	Transmit(ctx context.Context, report *MockedSignedReport, donID string) error
}

type overlayTransmitter struct {
	grpc              GrpcEndPoint
	reportManipulator func(*MockedSignedReport)
}

func NewOverlayTransmitter(grpc GrpcEndPoint, reportManipulator func(*MockedSignedReport)) OverlayTrasnmitter {
	return &overlayTransmitter{
		grpc:              grpc,
		reportManipulator: reportManipulator,
	}
}

func (t *overlayTransmitter) Transmit(ctx context.Context, report *MockedSignedReport, donID string) error {
	conn, err := t.grpc.Connect()
	if err != nil {
		return err
	}
	t.reportManipulator(report)
	data, err := MarshalReport(report)
	if err != nil {
		return err
	}
	client := proto.NewControlServiceClient(conn)
	_, err = client.Publish(ctx, &proto.PublishRequest{
		Topic: donTopic(donID),
		Data:  data,
	})
	return err
}
