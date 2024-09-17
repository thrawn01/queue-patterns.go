package queue

import (
	"context"
	pb "github.com/thrawn01/queue-patterns.go/proto"
	"sync"
	"time"
)

type Querator struct {
	requestCh  chan *Request
	wg         sync.WaitGroup
	done       chan struct{}
	client     *Client
	batchLimit int
}

func NewQuerator(limit int, c *Client) *Querator {
	ch := &Querator{
		requestCh:  make(chan *Request, limit),
		done:       make(chan struct{}),
		batchLimit: limit,
		client:     c,
	}

	ch.wg.Add(1)
	go ch.run()

	return ch
}

func (m *Querator) run() {
	defer m.wg.Done()

	for {
		select {
		// Collect all the requests into a local queue
		case req := <-m.requestCh:
			var batch pb.ProduceRequest
			requests := make([]*Request, 0, len(m.requestCh))
			batch.Items = make([]*pb.ProduceItem, 0, len(m.requestCh))
			batch.Items = append(batch.Items, req.Request.Items...)
			requests = append(requests, req)

		EMPTY:
			for {
				select {
				case req := <-m.requestCh:
					batch.Items = append(batch.Items, req.Request.Items...)
					requests = append(requests, req)
				default:
					break EMPTY
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := m.client.ProduceItems(ctx, &batch)
			for _, req := range requests {
				req.Err = err
				close(req.ReadyCh)
			}
			cancel()
		case <-m.done:
			return
		}
	}
}

func (m *Querator) Close(_ context.Context) error {
	close(m.done)
	m.wg.Wait()
	return nil
}

func (m *Querator) ProduceItems(ctx context.Context, req *pb.ProduceRequest) error {
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
