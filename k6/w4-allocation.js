/**
 * W4 — Memory / Allocation
 * VU levels: 10 / 50 / 100
 */
import http from "k6/http";
import { check } from "k6";
import { Trend, Rate, Counter } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const VU = parseInt(__ENV.VU || "10", 10);

const latency = new Trend("alloc_latency", true);
const errorRate = new Rate("alloc_errors");
const requests = new Counter("alloc_requests");

export const options = {
  stages: [
    { duration: "30s", target: VU },
    { duration: "60s", target: VU },
    { duration: "10s", target: 0 },
  ],
//  stages: [
//      { duration: "5s", target: VU },
//      { duration: "60s", target: VU },
//      { duration: "5s", target: 0 },
//    ],
  thresholds: {
    alloc_errors: ["rate<0.01"],
  },
};

const payload = JSON.stringify({
  objectCount: 10000,
  objectSizeBytes: 256,
  transformRounds: 3,
});

const params = {
  headers: { "Content-Type": "application/json" },
  tags: { name: "allocation" },
};

export default function () {
  const res = http.post(`${BASE_URL}/api/benchmark/allocation`, payload, params);

  const ok = check(res, {
    "status is 200": (r) => r.status === 200,
  });

  latency.add(res.timings.duration);
  errorRate.add(!ok);
  requests.add(1);
}
