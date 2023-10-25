package blstest

import "encoding/json"

type SignedReport struct {
	Network   string
	Data      []byte
	SigHex    string
	SeqNumber uint64
	SignerID  uint64
}

func MarshalSignedReport(r *SignedReport) ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalSignedReport(data []byte) (*SignedReport, error) {
	r := new(SignedReport)
	if err := json.Unmarshal(data, r); err != nil {
		return nil, err
	}
	return r, nil
}
