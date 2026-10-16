package com.benchmark.app.service;

import com.benchmark.app.model.Ticket;
import com.benchmark.app.repository.TicketRepository;
import org.springframework.stereotype.Service;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Random;

/**
 * Deterministic CPU / Allocation workloads for fair JVM vs Native comparison.
 */
@Service
public class BenchmarkService {

    private final TicketRepository ticketRepository;

    public BenchmarkService(TicketRepository ticketRepository) {
        this.ticketRepository = ticketRepository;
    }

    /**
     * W1 — CPU Bound
     * Fixed iterations, hash rounds, sort size → fully deterministic.
     */
    public CpuResult cpuBound(int iterations, int hashRounds, int sortSize) {
        long start = System.nanoTime();

        // 1. Hashing
        long hashAccumulator = 0;
        try {
            MessageDigest digest = MessageDigest.getInstance("SHA-256");
            for (int i = 0; i < iterations; i++) {
                byte[] input = ("benchmark-" + i).getBytes(StandardCharsets.UTF_8);
                for (int r = 0; r < hashRounds; r++) {
                    input = digest.digest(input);
                }
                hashAccumulator += input[0] & 0xFF;
            }
        } catch (NoSuchAlgorithmException e) {
            throw new IllegalStateException("SHA-256 not available", e);
        }

        // 2. Sorting
        int[] array = new int[sortSize];
        Random rnd = new Random(42); // fixed seed
        for (int i = 0; i < sortSize; i++) {
            array[i] = rnd.nextInt();
        }
        Arrays.sort(array);
        int median = array[sortSize / 2];

        // 3. Aggregation
        long sum = 0;
        for (int v : array) {
            sum += v;
        }

        long elapsedNs = System.nanoTime() - start;
        return new CpuResult(hashAccumulator, median, sum, elapsedNs);
    }

    /**
     * W2 — I/O Bound
     * Fixed query against PostgreSQL via Hikari.
     */
    public List<Ticket> ioBound(int limit) {
        return ticketRepository.findRecentForIoBenchmark(limit);
    }

    /**
     * W3 — High Concurrency (lightweight)
     * Mostly in-memory work + tiny DB touch to keep pool warm.
     */
    public ConcurrencyResult concurrencyBound() {
        long start = System.nanoTime();
        // Small deterministic work
        long acc = 0;
        for (int i = 0; i < 500; i++) {
            acc += (i * 31L) ^ (i << 3);
        }
        // Tiny DB hit so connection pool is exercised
        ticketRepository.findById(1L);
        long elapsedNs = System.nanoTime() - start;
        return new ConcurrencyResult(acc, elapsedNs);
    }

    /**
     * W4 — Allocation / Memory pressure
     * Controlled object allocation + temporary objects.
     */
    public AllocationResult allocationBound(int objectCount, int objectSizeBytes, int transformRounds) {
        long start = System.nanoTime();

        List<byte[]> objects = new ArrayList<>(objectCount);
        long totalBytes = 0;

        for (int i = 0; i < objectCount; i++) {
            byte[] data = new byte[objectSizeBytes];
            // Fill with deterministic pattern
            for (int j = 0; j < objectSizeBytes; j++) {
                data[j] = (byte) ((i + j) & 0xFF);
            }
            // Transform rounds (create temporary objects)
            for (int r = 0; r < transformRounds; r++) {
                byte[] transformed = new byte[objectSizeBytes];
                for (int j = 0; j < objectSizeBytes; j++) {
                    transformed[j] = (byte) (data[j] ^ (r + 1));
                }
                data = transformed;
            }
            objects.add(data);
            totalBytes += data.length;
        }

        // Force some retention so GC sees live set
        int checksum = 0;
        for (byte[] o : objects) {
            checksum += o[0] & 0xFF;
        }

        long elapsedNs = System.nanoTime() - start;
        return new AllocationResult(objectCount, totalBytes, checksum, elapsedNs);
    }

    // ── Result DTOs ──────────────────────────────────────────────────────────

    public record CpuResult(long hashAccumulator, int median, long sum, long elapsedNs) {}
    public record ConcurrencyResult(long accumulator, long elapsedNs) {}
    public record AllocationResult(int objectCount, long totalBytes, int checksum, long elapsedNs) {}
}
