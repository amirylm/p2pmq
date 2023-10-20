package don

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/amirylm/p2pmq/commons/utils"
)

type mockedDon struct {
	lock          sync.RWMutex
	threadControl utils.ThreadControl
	// DON ID
	id      string
	nodes   []*Node
	reports []MockedSignedReport
}

func newMockedDon(id string, nodes ...*Node) *mockedDon {
	for i, n := range nodes {
		n.Signers[id] = NewSigner(OracleID(i))
	}
	return &mockedDon{
		id:            id,
		nodes:         nodes,
		threadControl: utils.NewThreadControl(),
	}
}

func (d *mockedDon) Signers() map[OracleID]Signer {
	signers := map[OracleID]Signer{}
	for _, n := range d.nodes {
		s := n.Signers[d.id]
		signers[s.OracleID()] = s
	}
	return signers
}

func (d *mockedDon) run(interval time.Duration, subscribedDONs ...string) {
	d.threadControl.Go(func(ctx context.Context) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for _, n := range d.nodes {
			node := n
			if node == nil {
				// panic(fmt.Errorf("%s: node %d is nil", d.id, i))
				continue
			}
			d.threadControl.Go(func(c context.Context) {
				node.Start()
			})
			d.threadControl.Go(func(ctx context.Context) {
				if err := node.Consumer.Subscribe(ctx, subscribedDONs...); err != nil {
					if strings.Contains(err.Error(), "already tring to join") {
						return
					}
					panic(err)
				}
			})
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				d.broadcast(d.nextReport())
			}
		}
	})
}

func (d *mockedDon) stop() {
	d.threadControl.Close()
}

func (d *mockedDon) reportsCount() int {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return len(d.reports)
}

func (d *mockedDon) nextReport() *MockedSignedReport {
	d.lock.Lock()
	defer d.lock.Unlock()

	var lastSeq int64
	if len(d.reports) > 0 {
		lastReport := d.reports[len(d.reports)-1]
		lastSeq = lastReport.SeqNumber
	}

	r, err := NewMockedSignedReport(d.Signers(), lastSeq+1, d.id, []byte(fmt.Sprintf("dummy report #%d", lastSeq+1)))
	if err != nil {
		panic(err)
	}
	d.reports = append(d.reports, *r)
	return r
}

func (d *mockedDon) broadcast(r *MockedSignedReport) {
	for _, n := range d.nodes {
		node := n
		d.threadControl.Go(func(ctx context.Context) {
			// if err := d.signer.Verify(nil, r.Ctx, r.GetReportData(), r.Sig); err != nil {
			// 	if strings.Contains(err.Error(), "validation ignored") {
			// 		return
			// 	}
			// 	fmt.Printf("failed to verify report on don %s: %s\n", d.id, err)
			// 	return
			// }
			if err := node.Transmitter.Transmit(ctx, r, d.id); err != nil {
				if strings.Contains(err.Error(), "validation ignored") || ctx.Err() != nil {
					return
				}
				fmt.Printf("failed to publish report on don %s: %s\n", d.id, err)
			}
		})
	}
}

func wrapPanicErr(fn func(context.Context) error) func(context.Context) {
	return func(ctx context.Context) {
		err := fn(ctx)
		if err != nil {
			panic(err)
		}
	}
}
