/**
 * W2 — I/O Bound
 * VU levels: 10 / 50 / 100
 */
import http from "k6/http";
import { check } from "k6";
import { Trend, Rate, Counter } from "k6/metrics";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const VU = parseInt(__ENV.VU || "10", 10);

const latency = new Trend("io_latency", true);
const errorRate = new Rate("io_errors");
const requests = new Counter("io_requests");

export const options = {
  stages: [
    { duration: "30s", target: VU },
    { duration: "60s", target: VU },
    { duration: "10s", target: 0 },
  ],
  thresholds: {
    io_errors: ["rate<0.01"],
    http_req_duration: ["p(95)<500", "p(99)<1500"],
  },
};

export default function () {
  const res = http.get(`${BASE_URL}/api/benchmark/io?limit=20`, {
    tags: { name: "io" },
  });

  const ok = check(res, {
    "status is 200": (r) => r.status === 200,
    "has tickets": (r) => {
      try {
        return r.json("count") >= 0;
      } catch (_) {
        return false;
      }
    },
  });

  latency.add(res.timings.duration);
  errorRate.add(!ok);
  requests.add(1);
}
