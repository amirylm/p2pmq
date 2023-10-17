package donlib

import (
	"encoding/json"
	"fmt"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

func NewMockedSignedReport(signer Signer, seqNumber int64, srcDON string, data []byte) (*MockedSignedReport, error) {
	sr := &MockedSignedReport{
		SeqNumber: seqNumber,
		Src:       srcDON,
		Data:      data,
	}
	sig, err := signer.Sign(ocrtypes.ReportContext{}, []byte(fmt.Sprintf("%+v", sr)))
	if err != nil {
		return nil, err
	}
	sr.Sig = sig
	return sr, nil
}

type MockedSignedReport struct {
	SeqNumber int64
	// Src DON
	Src  string
	Data []byte
	Sig  []byte
}

func (r *MockedSignedReport) GetReportData() []byte {
	sr := &MockedSignedReport{
		SeqNumber: r.SeqNumber,
		Src:       r.Src,
		Data:      r.Data,
	}
	return []byte(fmt.Sprintf("%+v", sr))
}

func MarshalReport(r *MockedSignedReport) ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalReport(data []byte) (*MockedSignedReport, error) {
	r := new(MockedSignedReport)
	if err := json.Unmarshal(data, r); err != nil {
		return nil, err
	}
	return r, nil
}
