package don

import (
	"sync"
)

type ReportBuffer struct {
	lock    sync.RWMutex
	reports map[string][]MockedSignedReport

	bufferSize int
}

func NewReportBuffer(bufferSize int) *ReportBuffer {
	return &ReportBuffer{
		bufferSize: bufferSize,
	}
}

func (rb *ReportBuffer) Add(don string, sr MockedSignedReport) bool {
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

func (rb *ReportBuffer) Get(don string, seq int64) *MockedSignedReport {
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

func (rb *ReportBuffer) GetLatest(don string) *MockedSignedReport {
	rb.lock.RLock()
	defer rb.lock.RUnlock()

	if len(rb.reports) == 0 {
		return nil
	}
	buf := rb.reports[don]
	if len(buf) == 0 {
		return nil
	}

	var latest MockedSignedReport
	for _, r := range buf {
		if r.SeqNumber > latest.SeqNumber {
			latest = r
		}
	}
	return &latest
}
