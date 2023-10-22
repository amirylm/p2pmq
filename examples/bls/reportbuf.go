package blstest

import (
	"sync"

	"github.com/amirylm/p2pmq/proto"
)

var (
	reportBufferSize = 128

	skipThreshold    uint64 = 5
	invalidThreshold uint64 = 25
)

type ReportBuffer struct {
	lock    sync.RWMutex
	reports map[string][]SignedReport

	bufferSize int
}

func NewReportBuffer(bufferSize int) *ReportBuffer {
	return &ReportBuffer{
		bufferSize: bufferSize,
	}
}

func (rb *ReportBuffer) Add(net string, sr SignedReport) bool {
	rb.lock.Lock()
	defer rb.lock.Unlock()

	if rb.reports == nil {
		rb.reports = make(map[string][]SignedReport)
	}
	buf := rb.reports[net]
	if len(buf) == 0 {
		buf = make([]SignedReport, rb.bufferSize)
	}
	i := sr.SeqNumber % uint64(rb.bufferSize)
	// update only if newer than existing report in buffer
	if buf[i].SeqNumber < sr.SeqNumber {
		buf[i] = sr
		rb.reports[net] = buf
		return true
	}
	return false
}

func (rb *ReportBuffer) Get(net string, seq uint64) *SignedReport {
	rb.lock.RLock()
	defer rb.lock.RUnlock()

	if rb.reports == nil {
		return nil
	}
	buf := rb.reports[net]
	if len(buf) == 0 {
		return nil
	}
	i := seq % uint64(rb.bufferSize)
	if buf[i].SeqNumber == seq {
		r := buf[i]
		return &r
	}
	return nil
}

func (rb *ReportBuffer) Lookup(seq uint64) []string {
	if len(rb.reports) == 0 {
		return nil
	}
	var ids []string
	i := seq % uint64(rb.bufferSize)
	for id, buf := range rb.reports {
		if buf[i].SeqNumber == seq {
			ids = append(ids, id)
		}
	}

	return ids
}

func (rb *ReportBuffer) ValidateSequence(sr SignedReport) proto.ValidationResult {
	latest := rb.getLatest(sr.Network)
	if latest != nil {
		diff := sr.SeqNumber - latest.SeqNumber
		switch {
		case diff > invalidThreshold:
			return proto.ValidationResult_REJECT
		case diff > skipThreshold:
			return proto.ValidationResult_IGNORE
		default: // less than skipThreshold, accept
		}
	}
	if rb.Get(sr.Network, sr.SeqNumber) != nil {
		// ignore if this report was seen already
		return proto.ValidationResult_IGNORE
	}
	return proto.ValidationResult_ACCEPT
}

func (rb *ReportBuffer) getLatest(net string) *SignedReport {
	rb.lock.RLock()
	defer rb.lock.RUnlock()

	if len(rb.reports) == 0 {
		return nil
	}
	buf := rb.reports[net]
	if len(buf) == 0 {
		return nil
	}

	var latest SignedReport
	for _, r := range buf {
		if r.SeqNumber > latest.SeqNumber {
			latest = r
		}
	}
	return &latest
}
