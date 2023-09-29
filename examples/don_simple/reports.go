package donsimple

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/polydawn/refmt/json"
)

type reportsBuffer struct {
	lock    sync.RWMutex
	reports map[string][]MockedSignedReport

	bufferSize int
}

func (rb *reportsBuffer) addReport(don string, sr MockedSignedReport) bool {
	rb.lock.Lock()
	defer rb.lock.Unlock()

	if rb.reports == nil {
		rb.reports = make(map[string][]MockedSignedReport)
	}
	buf := rb.reports[don]
	if len(buf) == 0 {
		buf = make([]MockedSignedReport, rb.bufferSize)
	}
	i := sr.SeqNumber % int64(rb.bufferSize)
	// update only if newer than existing report in buffer
	if buf[i].SeqNumber < sr.SeqNumber {
		buf[i] = sr
		rb.reports[don] = buf
		return true
	}
	return false
}

func (rb *reportsBuffer) getReport(don string, seq int64) *MockedSignedReport {
	rb.lock.RLock()
	defer rb.lock.RUnlock()

	if rb.reports == nil {
		return nil
	}
	buf := rb.reports[don]
	if len(buf) == 0 {
		return nil
	}

	i := seq % int64(rb.bufferSize)
	if buf[i].SeqNumber == seq {
		r := buf[i]
		return &r
	}
	return nil
}

func newSignedReport(seqNumber int64, srcDON string, data []byte) *MockedSignedReport {
	sr := &MockedSignedReport{
		SeqNumber: seqNumber,
		Src:       srcDON,
		Data:      data,
	}
	// mocked, using sha256 as signature
	h := sha256.Sum256([]byte(fmt.Sprintf("%+v", sr)))
	sr.Sig = h[:]
	return sr
}

type MockedSignedReport struct {
	SeqNumber int64
	// Src DON
	Src  string
	Data []byte
	Sig  []byte
}

// mocked, using sha256 as signature
func (sr *MockedSignedReport) verify() error {
	signed := newSignedReport(sr.SeqNumber, sr.Src, sr.Data)
	if !bytes.Equal(sr.Sig, signed.Sig) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
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
