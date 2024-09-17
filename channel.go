package queue

import (
	"context"
	"github.com/thrawn01/queue-patterns.go/interval"
	pb "github.com/thrawn01/queue-patterns.go/proto"
	"sync"
	"time"
)

type Channel struct {
	requestCh  chan *Request
	wg         sync.WaitGroup
	done       chan struct{}
	client     *Client
	batchLimit int
}

func NewChannel(limit int, c *Client) *Channel {
	ch := &Channel{
		requestCh:  make(chan *Request, limit),
		done:       make(chan struct{}),
		batchLimit: limit,
		client:     c,
	}

	ch.wg.Add(1)
	go ch.run()

	return ch
}

func (m *Channel) run() {
	defer m.wg.Done()
	queue := make([]*Request, 0, m.batchLimit)

	i := interval.NewInterval(15 * time.Millisecond)
	i.Next()

	for {
		select {
		// Collect all the requests into a local queue
		case req := <-m.requestCh:
			queue = append(queue, req)

		// Once every tick send all the requests in a batch
		case <-i.C:
			var batch pb.ProduceRequest
			for _, req := range queue {
				batch.Items = append(batch.Items, req.Request.Items...)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := m.client.ProduceItems(ctx, &batch)
			for _, req := range queue {
				req.Err = err
				close(req.ReadyCh)
			}
			queue = make([]*Request, 0, m.batchLimit)
			cancel()
			i.Next()
		case <-m.done:
			return
		}
	}
}

func (m *Channel) Close(_ context.Context) error {
	close(m.done)
	m.wg.Wait()
	return nil
}

func (m *Channel) ProduceItems(ctx context.Context, req *pb.ProduceRequest) error {
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
