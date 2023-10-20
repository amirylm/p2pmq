package don

import (
	"encoding/json"
	"fmt"
)

func NewMockedSignedReport(signers map[OracleID]Signer, seqNumber int64, srcDON string, data []byte) (*MockedSignedReport, error) {
	sr := &MockedSignedReport{
		SeqNumber: seqNumber,
		Src:       srcDON,
		Data:      data,
	}
	rctx := ReportContext{
		ReportTimestamp: ReportTimestamp{
			ConfigDigest: ConfigDigest{},
		},
	}

	sr.Sigs = make(map[OracleID][]byte)
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
	Sigs      map[OracleID][]byte
	Ctx       ReportContext
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
