/**
 * W3 — High Concurrency
 * Levels: 10 → 50 → 100 → 200
 */
import http from "k6/http";
import { check } from "k6";
import { Trend, Rate, Counter } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const VU = parseInt(__ENV.VU || "10", 10);

const latency = new Trend("concurrency_latency", true);
const errorRate = new Rate("concurrency_errors");
const requests = new Counter("concurrency_requests");

export const options = {
  stages: [
    { duration: "30s", target: VU },
    { duration: "60s", target: VU },
    { duration: "10s", target: 0 },
  ],
  thresholds: {
    concurrency_errors: ["rate<0.02"],
  },
};

export default function () {
  const res = http.get(`${BASE_URL}/api/benchmark/concurrency`, {
    tags: { name: "concurrency" },
  });

  const ok = check(res, {
    "status is 200": (r) => r.status === 200,
  });

  latency.add(res.timings.duration);
  errorRate.add(!ok);
  requests.add(1);
}
