package com.benchmark.app.controller;

import com.benchmark.app.model.Ticket;
import com.benchmark.app.service.BenchmarkService;
import com.benchmark.app.service.BenchmarkService.AllocationResult;
import com.benchmark.app.service.BenchmarkService.ConcurrencyResult;
import com.benchmark.app.service.BenchmarkService.CpuResult;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/benchmark")
public class BenchmarkController {

    private final BenchmarkService benchmarkService;

    public BenchmarkController(BenchmarkService benchmarkService) {
        this.benchmarkService = benchmarkService;
    }

    /**
     * W1 — CPU Bound
     * Body example:
     * {
     *   "iterations": 10000,
     *   "hashRounds": 5,
     *   "sortSize": 5000
     * }
     */
    @PostMapping("/cpu")
    public ResponseEntity<Map<String, Object>> cpu(@RequestBody Map<String, Integer> body) {
        int iterations = body.getOrDefault("iterations", 10000);
        int hashRounds = body.getOrDefault("hashRounds", 5);
        int sortSize   = body.getOrDefault("sortSize", 5000);

        CpuResult result = benchmarkService.cpuBound(iterations, hashRounds, sortSize);

        Map<String, Object> response = new HashMap<>();
        response.put("hashAccumulator", result.hashAccumulator());
        response.put("median", result.median());
        response.put("sum", result.sum());
        response.put("elapsedNs", result.elapsedNs());
        response.put("elapsedMs", result.elapsedNs() / 1_000_000.0);
        return ResponseEntity.ok(response);
    }

    /**
     * W2 — I/O Bound
     * Fixed query against PostgreSQL.
     */
    @GetMapping("/io")
    public ResponseEntity<Map<String, Object>> io(
            @RequestParam(defaultValue = "20") int limit) {
        List<Ticket> tickets = benchmarkService.ioBound(Math.min(limit, 100));
        Map<String, Object> response = new HashMap<>();
        response.put("count", tickets.size());
        response.put("tickets", tickets);
        return ResponseEntity.ok(response);
    }

    /**
     * W3 — High Concurrency
     * Lightweight work + tiny DB touch.
     */
    @GetMapping("/concurrency")
    public ResponseEntity<Map<String, Object>> concurrency() {
        ConcurrencyResult result = benchmarkService.concurrencyBound();
        Map<String, Object> response = new HashMap<>();
        response.put("accumulator", result.accumulator());
        response.put("elapsedNs", result.elapsedNs());
        return ResponseEntity.ok(response);
    }

    /**
     * W4 — Allocation / Memory
     * Body example:
     * {
     *   "objectCount": 10000,
     *   "objectSizeBytes": 256,
     *   "transformRounds": 3
     * }
     */
    @PostMapping("/allocation")
    public ResponseEntity<Map<String, Object>> allocation(@RequestBody Map<String, Integer> body) {
        int objectCount      = body.getOrDefault("objectCount", 10000);
        int objectSizeBytes  = body.getOrDefault("objectSizeBytes", 256);
        int transformRounds  = body.getOrDefault("transformRounds", 3);

        // Safety caps to avoid OOM during benchmark
        objectCount = Math.min(objectCount, 50_000);
        objectSizeBytes = Math.min(objectSizeBytes, 1024);

        AllocationResult result = benchmarkService.allocationBound(objectCount, objectSizeBytes, transformRounds);

        Map<String, Object> response = new HashMap<>();
        response.put("objectCount", result.objectCount());
        response.put("totalBytes", result.totalBytes());
        response.put("totalMB", result.totalBytes() / (1024.0 * 1024.0));
        response.put("checksum", result.checksum());
        response.put("elapsedNs", result.elapsedNs());
        response.put("elapsedMs", result.elapsedNs() / 1_000_000.0);
        return ResponseEntity.ok(response);
    }
}
