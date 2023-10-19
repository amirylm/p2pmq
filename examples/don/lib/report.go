package donlib

import (
	"encoding/json"
	"fmt"

	"github.com/smartcontractkit/libocr/commontypes"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

func NewMockedSignedReport(signers map[commontypes.OracleID]Signer, seqNumber int64, srcDON string, data []byte) (*MockedSignedReport, error) {
	sr := &MockedSignedReport{
		SeqNumber: seqNumber,
		Src:       srcDON,
		Data:      data,
	}
	rctx := ocrtypes.ReportContext{
		ReportTimestamp: ocrtypes.ReportTimestamp{
			ConfigDigest: ocrtypes.ConfigDigest{},
		},
	}

	sr.Sigs = make(map[commontypes.OracleID][]byte)
	for oid, s := range signers {
		sig, err := s.Sign(rctx, []byte(fmt.Sprintf("%+v", sr)))
		if err != nil {
			return nil, err
		}
		sr.Sigs[oid] = sig
	}

	sr.Ctx = rctx
	return sr, nil
}

type MockedSignedReport struct {
	// Src DON
	Src       string
	SeqNumber int64
	Data      []byte
	Sigs      map[commontypes.OracleID][]byte
	Ctx       ocrtypes.ReportContext
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
