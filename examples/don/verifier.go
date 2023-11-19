package don

import (
	"fmt"

	"github.com/amirylm/p2pmq/proto"
)

var (
	skipThreshold    int64 = 5
	invalidThreshold int64 = 25
)

// reportVerifier is doing the actual validation on incoming messages
type reportVerifier struct {
	reports *ReportBuffer
	dons    map[string]map[OracleID]OnchainPublicKey
}

func NewReportVerifier(reports *ReportBuffer) *reportVerifier {
	return &reportVerifier{
		reports: reports,
		dons:    make(map[string]map[OracleID]OnchainPublicKey),
	}
}

func (rv *reportVerifier) Process(msg *proto.Message) ([]byte, proto.ValidationResult) {
	raw := msg.GetData()
	r, err := UnmarshalReport(raw)
	if err != nil || r == nil {
		// bad encoding
		return raw, proto.ValidationResult_REJECT
	}
	pubkeys, ok := rv.dons[r.Src]
	if !ok {
		return raw, proto.ValidationResult_IGNORE
	}

	s := NewSigner(0)

	valid := 0
	for oid, pk := range pubkeys {
		sig, ok := r.Sigs[oid]
		if !ok {
			continue
		}
		if err := s.Verify(pk, r.Ctx, r.GetReportData(), sig); err != nil {
			fmt.Printf("failed to verify report: %v\n", err)
			return raw, proto.ValidationResult_REJECT
		}
		valid++
	}

	// n = 3f + 1
	// n-1 = 3f
	// f = (n-1)/3
	n := len(pubkeys)
	f := (n - 1) / 3
	threshold := n - f
	if valid < threshold {
		return raw, proto.ValidationResult_REJECT
	}

	return raw, rv.validateSequence(r)
}

func (rv *reportVerifier) validateSequence(r *MockedSignedReport) proto.ValidationResult {
	latest := rv.reports.GetLatest(r.Src)
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
	if rv.reports.Get(r.Src, r.SeqNumber) != nil {
		return proto.ValidationResult_IGNORE
	}
	return proto.ValidationResult_ACCEPT
}
