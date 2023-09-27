package tests

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/amirylm/p2pmq/commons/utils"
	"github.com/amirylm/p2pmq/proto"
)

type mockedDon struct {
	// DON ID
	id      string
	nodes   []*nodeClient
	reports []MockedSignedReport

	lock          sync.RWMutex
	threadControl utils.ThreadControl
}

func newDon(id string, nodes ...*nodeClient) *mockedDon {
	return &mockedDon{
		id:            id,
		nodes:         nodes,
		threadControl: utils.NewThreadControl(),
	}
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
				if err := node.subscribe(c, subscribedDONs...); err != nil {
					panic(err)
				}
			})
			d.threadControl.Go(func(c context.Context) {
				if err := node.listen(c); err != nil {
					panic(err)
				}
			})
			d.threadControl.Go(func(c context.Context) {
				if err := node.validate(c); err != nil {
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
	r := newSignedReport(lastSeq+1, d.id, []byte(fmt.Sprintf("dummy report #%d", lastSeq+1)))
	d.reports = append(d.reports, *r)
	return r
}

func (d *mockedDon) broadcast(r *MockedSignedReport) {
	for _, n := range d.nodes {
		n.reports.addReport(d.id, *r)
		d.threadControl.Go(func(ctx context.Context) {
			conn, err := n.connect(true)
			if err != nil {
				return
			}
			data, err := MarshalReport(r)
			if err != nil {
				return
			}
			controlClient := proto.NewControlServiceClient(conn)
			_, err = controlClient.Publish(ctx, &proto.PublishRequest{
				Topic: fmt.Sprintf("don.%s", d.id),
				Data:  data,
			})
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
