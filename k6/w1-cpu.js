/**
 * W1 — CPU Bound
 * VU levels: 1 / 10 / 50 / 100
 * Duration: 60s measurement after 30s warm-up
 */
import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Rate, Counter } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const VU = parseInt(__ENV.VU || "10", 10);

const latency = new Trend("cpu_latency", true);
const errorRate = new Rate("cpu_errors");
const requests = new Counter("cpu_requests");

export const options = {
  stages: [
    { duration: "30s", target: VU }, // warm-up
    { duration: "60s", target: VU }, // measurement
    { duration: "10s", target: 0 },  // ramp-down
  ],
  thresholds: {
    cpu_errors: ["rate<0.01"],
    http_req_duration: ["p(95)<2000", "p(99)<5000"],
  },
};

const payload = JSON.stringify({
  iterations: 10000,
  hashRounds: 5,
  sortSize: 5000,
});

const params = {
  headers: { "Content-Type": "application/json" },
  tags: { name: "cpu" },
};

export default function () {
  const res = http.post(`${BASE_URL}/api/benchmark/cpu`, payload, params);

  const ok = check(res, {
    "status is 200": (r) => r.status === 200,
  });

  latency.add(res.timings.duration);
  errorRate.add(!ok);
  requests.add(1);

  // No think time — pure CPU pressure
}
