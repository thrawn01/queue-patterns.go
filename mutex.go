package queue

import (
	"context"
	"github.com/thrawn01/queue-patterns.go/interval"
	pb "github.com/thrawn01/queue-patterns.go/proto"
	"sync"
	"time"
)

type Mutex struct {
	queue      []*Request
	wg         sync.WaitGroup
	done       chan struct{}
	mutex      sync.Mutex
	client     *Client
	batchLimit int
}

func NewMutex(limit int, c *Client) *Mutex {
	m := &Mutex{
		queue:      make([]*Request, 0, limit),
		done:       make(chan struct{}),
		batchLimit: limit,
		client:     c,
	}

	m.wg.Add(1)
	go m.run()

	return m
}

func (m *Mutex) sendQueue() {
	var batch pb.ProduceRequest
	for _, req := range m.queue {
		batch.Items = append(batch.Items, req.Request.Items...)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := m.client.ProduceItems(ctx, &batch)
	for _, req := range m.queue {
		req.Err = err
		close(req.ReadyCh)
	}
	m.queue = make([]*Request, 0, m.batchLimit)
}

func (m *Mutex) ProduceItems(ctx context.Context, req *pb.ProduceRequest) error {
	m.mutex.Lock()
	r := Request{
		ReadyCh: make(chan struct{}),
		Request: req,
		Context: ctx,
	}
	m.queue = append(m.queue, &r)
	if len(m.queue) >= m.batchLimit {
		m.sendQueue()
	}
	m.mutex.Unlock()

	select {
	case <-r.ReadyCh:
		return r.Err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *Mutex) run() {
	defer m.wg.Done()

	// TODO: Experiment with the interval, use a set tick instead, similar to tiger beetle?
	i := interval.NewInterval(15 * time.Millisecond)
	i.Next()

	for {
		select {
		case <-i.C:
			m.mutex.Lock()
			m.sendQueue()
			m.mutex.Unlock()
			i.Next()
		case <-m.done:
			return
		}
	}
}

func (m *Mutex) Close(_ context.Context) error {
	close(m.done)
	m.wg.Wait()
	return nil
}
