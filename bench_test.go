package queue_test

import (
	"context"
	"fmt"
	"github.com/kapetan-io/tackle/clock"
	"github.com/kapetan-io/tackle/random"
	"github.com/stretchr/testify/require"
	"github.com/thrawn01/queue-patterns.go"
	pb "github.com/thrawn01/queue-patterns.go/proto"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkQueuePatterns(b *testing.B) {
	s, err := queue.NewServer(context.Background(), queue.Config{
		ListenAddress: "localhost:2319",
	})
	require.NoError(b, err)
	c := s.MustClient()

	items := generateProduceItems(1_000)
	mask := len(items) - 1

	b.Run("noQueue", func(b *testing.B) {
		start := clock.Now()
		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			index := int(rand.Uint32() & uint32(mask))
			for p.Next() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				if err := c.ProduceItems(ctx, &pb.ProduceRequest{
					Items: items[index&mask : index+1&mask],
				}); err != nil {
					b.Error(err)
				}
				cancel()
			}
		})
		opsPerSec := float64(b.N) / clock.Since(start).Seconds()
		b.ReportMetric(opsPerSec, "ops/s")
	})

	b.Run("mutexQueue", func(b *testing.B) {
		m := queue.NewMutex(1_000, c)

		start := clock.Now()
		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			index := int(rand.Uint32() & uint32(mask))
			for p.Next() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				if err := m.ProduceItems(ctx, &pb.ProduceRequest{
					Items: items[index&mask : index+1&mask],
				}); err != nil {
					b.Error(err)
				}
				cancel()
			}
		})
		require.NoError(b, m.Close(context.Background()))
		opsPerSec := float64(b.N) / clock.Since(start).Seconds()
		b.ReportMetric(opsPerSec, "ops/s")
	})
}

func generateProduceItems(size int) []*pb.ProduceItem {
	items := make([]*pb.ProduceItem, 0, size)
	for i := 0; i < size; i++ {
		items = append(items, &pb.ProduceItem{
			Bytes: []byte(fmt.Sprintf("%d-%s", i, random.String("payload-", 256))),
		})
	}
	return items
}
