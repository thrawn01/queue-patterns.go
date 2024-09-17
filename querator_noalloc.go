package queue

import (
	"context"
	pb "github.com/thrawn01/queue-patterns.go/proto"
	"sync"
	"time"
)

type QueratorNoAlloc struct {
	requestCh  chan *Request
	wg         sync.WaitGroup
	done       chan struct{}
	client     *Client
	batchLimit int
}

func NewQueratorNoAlloc(limit int, c *Client) *QueratorNoAlloc {
	ch := &QueratorNoAlloc{
		requestCh:  make(chan *Request, limit*10),
		done:       make(chan struct{}),
		batchLimit: limit,
		client:     c,
	}

	ch.wg.Add(1)
	go ch.run()

	return ch
}

func (m *QueratorNoAlloc) run() {
	defer m.wg.Done()
	var batch pb.ProduceRequest
	requests := make([]*Request, 10_000)
	batch.Items = make([]*pb.ProduceItem, 0, 10_000)
	var idx int

	for {
		select {
		// Collect all the requests into a local queue
		case req := <-m.requestCh:
			batch.Items = append(batch.Items, req.Request.Items...)
			requests[idx] = req
			idx++

		EMPTY:
			for {
				select {
				case req := <-m.requestCh:
					batch.Items = append(batch.Items, req.Request.Items...)
					requests[idx] = req
					idx++
				default:
					break EMPTY
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := m.client.ProduceItems(ctx, &batch)
			for i := 0; i < idx; i++ {
				requests[i].Err = err
				close(requests[i].ReadyCh)
			}
			cancel()
			batch.Items = batch.Items[:0]
			//requests = requests[:0]
			idx = 0
		case <-m.done:
			return
		}
	}
}

func (m *QueratorNoAlloc) Close(_ context.Context) error {
	close(m.done)
	m.wg.Wait()
	return nil
}

func (m *QueratorNoAlloc) ProduceItems(ctx context.Context, req *pb.ProduceRequest) error {
	r := Request{
		ReadyCh: make(chan struct{}),
		Request: req,
		Context: ctx,
	}

	m.requestCh <- &r

	select {
	case <-r.ReadyCh:
		return r.Err
	case <-ctx.Done():
		return ctx.Err()
	}
}
