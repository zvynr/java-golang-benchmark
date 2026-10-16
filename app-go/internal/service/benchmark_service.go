package service

import (
	"benchmark-go/internal/model"
	"benchmark-go/internal/pool"
	"benchmark-go/internal/repository"
	"benchmark-go/internal/worker"
	"context"
	"crypto/sha256"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

type BenchmarkService interface {
	CpuBound(ctx context.Context, iterations, hashRounds, sortSize int) (*model.CpuResult, error)
	IoBound(ctx context.Context, limit int) ([]model.Ticket, error)
	ConcurrencyBound(ctx context.Context) (*model.ConcurrencyResult, error)
	AllocationBound(ctx context.Context, objectCount, objectSizeBytes, transformRounds int) (*model.AllocationResult, error)
}

type benchmarkService struct {
	ticketRepo repository.TicketRepository
	poolMgr    *worker.PoolManager
}

func NewBenchmarkService(ticketRepo repository.TicketRepository, poolMgr *worker.PoolManager) BenchmarkService {
	return &benchmarkService{
		ticketRepo: ticketRepo,
		poolMgr:    poolMgr,
	}
}

func (s *benchmarkService) CpuBound(ctx context.Context, iterations, hashRounds, sortSize int) (*model.CpuResult, error) {
	if err := s.poolMgr.CpuPool.Acquire(ctx); err != nil {
		return nil, err
	}
	defer s.poolMgr.CpuPool.Release()

	start := time.Now()

	// 1. Hashing (Sequential per request - protected by Dedicated CPU Worker Pool)
	var hashAccumulator int64
	var (
		hashSum [sha256.Size]byte
		tmp     [32]byte
	)

	for i := 0; i < iterations; i++ {
		if i%1000 == 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}

		input := append(tmp[:0], "benchmark-"...)
		input = strconv.AppendInt(input, int64(i), 10)

		for r := 0; r < hashRounds; r++ {
			hashSum = sha256.Sum256(input)
			input = hashSum[:]
		}
		hashAccumulator += int64(input[0])
	}

	// 2. Sorting
	arr := make([]int, sortSize)
	rnd := rand.New(rand.NewSource(42))
	for i := 0; i < sortSize; i++ {
		arr[i] = rnd.Int()
	}
	sort.Ints(arr)
	median := arr[sortSize/2]

	// 3. Aggregation
	var sum int64
	for _, v := range arr {
		sum += int64(v)
	}

	elapsedNs := time.Since(start).Nanoseconds()
	return &model.CpuResult{
		HashAccumulator: hashAccumulator,
		Median:          median,
		Sum:             sum,
		ElapsedNs:       elapsedNs,
		ElapsedMs:       float64(elapsedNs) / 1000000.0,
	}, nil
}

func (s *benchmarkService) IoBound(ctx context.Context, limit int) ([]model.Ticket, error) {
	if limit > 100 {
		limit = 100
	}
	return s.ticketRepo.FindRecentForIoBenchmark(ctx, limit)
}

func (s *benchmarkService) ConcurrencyBound(ctx context.Context) (*model.ConcurrencyResult, error) {
    start := time.Now()
    var acc int64
    for i := 0; i < 500; i++ {
        acc += (int64(i) * 31) ^ (int64(i) << 3)
    }

    // DB I/O workload (sequential)
    _, err := s.ticketRepo.FindByID(ctx, 1)
    if err != nil {
        return nil, err
    }

    elapsed := time.Since(start)

    return &model.ConcurrencyResult{
        Accumulator: acc,
        ElapsedNs:   elapsed.Nanoseconds(),
    }, nil
}

func (s *benchmarkService) AllocationBound(ctx context.Context, objectCount, objectSizeBytes, transformRounds int) (*model.AllocationResult, error) {
	if err := s.poolMgr.AllocPool.Acquire(ctx); err != nil {
		return nil, err
	}
	defer s.poolMgr.AllocPool.Release()

	start := time.Now()

	if objectCount > 50000 {
		objectCount = 50000
	}
	if objectSizeBytes > 1024 {
		objectSizeBytes = 1024
	}

	objects := make([][]byte, 0, objectCount)
	var totalBytes int64

	// Protected by Dedicated Alloc Worker Pool, with immediate sync.Pool recycling!
	for i := 0; i < objectCount; i++ {
		if i%1000 == 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}

		var data []byte
		var pSlice *[]byte
		if objectSizeBytes == 256 {
			pSlice = pool.GetByteSlice256()
			data = *pSlice
		} else {
			data = make([]byte, objectSizeBytes)
		}

		for j := 0; j < objectSizeBytes; j++ {
			data[j] = byte((i + j) & 0xFF)
		}

		for r := 0; r < transformRounds; r++ {
			transformed := make([]byte, objectSizeBytes)
			for j := 0; j < objectSizeBytes; j++ {
				transformed[j] = data[j] ^ byte(r+1)
			}
			data = transformed
		}

		objects = append(objects, data)
		totalBytes += int64(len(data))

		if pSlice != nil {
			pool.PutByteSlice256(pSlice)
		}
	}

	var checksum int
	for _, o := range objects {
		if len(o) > 0 {
			checksum += int(o[0])
		}
	}

	elapsedNs := time.Since(start).Nanoseconds()
	return &model.AllocationResult{
		ObjectCount: objectCount,
		TotalBytes:  totalBytes,
		TotalMB:     float64(totalBytes) / (1024.0 * 1024.0),
		Checksum:    checksum,
		ElapsedNs:   elapsedNs,
		ElapsedMs:   float64(elapsedNs) / 1000000.0,
	}, nil
}
