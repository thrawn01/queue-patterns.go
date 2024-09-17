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
	"runtime"
	"testing"
	"time"
)

func BenchmarkQueuePatterns(b *testing.B) {
	fmt.Printf("Current Operating System has '%d' CPUs\n", runtime.NumCPU())
	s, err := queue.NewServer(context.Background(), queue.Config{
		ListenAddress: "localhost:2319",
		RequestSleep:  10 * time.Millisecond,
	})
	require.NoError(b, err)
	c := s.MustClient()

	items := generateProduceItems(1_000)
	mask := len(items) - 1

	b.Run("none", func(b *testing.B) {
		start := clock.Now()
		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			index := int(rand.Uint32() & uint32(mask))
			for p.Next() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	b.Run("mutex", func(b *testing.B) {
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

	b.Run("channel", func(b *testing.B) {
		ch := queue.NewChannel(1_000, c)

		start := clock.Now()
		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			index := int(rand.Uint32() & uint32(mask))
			for p.Next() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				if err := ch.ProduceItems(ctx, &pb.ProduceRequest{
					Items: items[index&mask : index+1&mask],
				}); err != nil {
					b.Error(err)
				}
				cancel()
			}
		})
		require.NoError(b, ch.Close(context.Background()))
		opsPerSec := float64(b.N) / clock.Since(start).Seconds()
		b.ReportMetric(opsPerSec, "ops/s")
	})

	b.Run("querator", func(b *testing.B) {
		q := queue.NewQuerator(1_000, c)

		start := clock.Now()
		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			index := int(rand.Uint32() & uint32(mask))
			for p.Next() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				if err := q.ProduceItems(ctx, &pb.ProduceRequest{
					Items: items[index&mask : index+1&mask],
				}); err != nil {
					b.Error(err)
				}
				cancel()
			}
		})
		require.NoError(b, q.Close(context.Background()))
		opsPerSec := float64(b.N) / clock.Since(start).Seconds()
		b.ReportMetric(opsPerSec, "ops/s")
	})

	b.Run("querator-noalloc", func(b *testing.B) {
		q := queue.NewQueratorNoAlloc(1_000, c)

		start := clock.Now()
		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			index := int(rand.Uint32() & uint32(mask))
			for p.Next() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				if err := q.ProduceItems(ctx, &pb.ProduceRequest{
					Items: items[index&mask : index+1&mask],
				}); err != nil {
					b.Error(err)
				}
				cancel()
			}
		})
		require.NoError(b, q.Close(context.Background()))
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
