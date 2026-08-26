package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"epochclock/internal/alloc"
	"epochclock/internal/clock"
	"epochclock/internal/lease"
	"epochclock/internal/metric"
	"epochclock/internal/node"
	"epochclock/internal/persist"
)

func main() {
	var (
		addr        = flag.String("addr", ":8080", "HTTP 监听地址")
		dataDir     = flag.String("data", "./data", "WAL 数据目录")
		leaseTTL    = flag.Duration("lease-ttl", 30*time.Second, "签发租约有效期")
		batchSize   = flag.Uint64("batch", 100, "默认签发批次大小")
		segmentSize = flag.Int64("segment", 64*1024, "WAL 段文件大小")
	)
	flag.Parse()

	metrics := metric.New()
	nodes := node.NewRegistry()
	seed := nodes.Register("seed-node")
	log.Printf("seed node registered: %s (%s)", seed.ID, seed.Name)

	wall := clock.SystemWall()
	hlc := clock.New(0, 0)
	leases := lease.NewManager(nodes, time.Now, metrics)

	pers, err := persist.NewWALPersister(*dataDir, *segmentSize)
	if err != nil {
		log.Fatalf("open wal: %v", err)
	}
	defer func() {
		if err := pers.Close(); err != nil {
			log.Printf("close wal: %v", err)
		}
	}()

	result, err := pers.Replay()
	if err != nil {
		log.Fatalf("replay wal: %v", err)
	}
	metrics.AddReplayApplied(uint64(result.Applied))
	metrics.AddReplaySkipped(uint64(result.Skipped))
	log.Printf("replay complete: %d ranges applied, %d skipped, clock=%s",
		result.Applied, result.Skipped, result.Clock.String())

	allocator := alloc.NewAllocator(pers, hlc, leases, metrics, wall, *batchSize)
	if err := allocator.Restore(result); err != nil {
		log.Fatalf("restore allocator: %v", err)
	}
	seedLease, err := leases.Grant(seed.ID, *leaseTTL)
	if err != nil {
		log.Fatalf("grant seed lease: %v", err)
	}
	log.Printf("seed lease granted: %s holder=%s ttl=%s", seedLease.ID, seedLease.NodeID, *leaseTTL)

	server := NewServer(ServerDeps{
		Allocator: allocator,
		Leases:    leases,
		Nodes:     nodes,
		HLC:       hlc,
		Metrics:   metrics,
		Persister: pers,
		StartedAt: time.Now().UTC(),
	})
	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("epochclock listening on %s", *addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Printf("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
