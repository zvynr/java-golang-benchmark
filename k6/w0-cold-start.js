/**
 * W0 — Cold Start
 * Measures time-to-ready by polling /actuator/health until UP.
 * Run against a freshly started container.
 */
import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 1,
  duration: "120s",
  thresholds: {
    checks: ["rate>0.99"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

export default function () {
  const res = http.get(`${BASE_URL}/actuator/health`, {
    tags: { name: "health" },
  });

  const ok = check(res, {
    "status is 200": (r) => r.status === 200,
    "status is UP": (r) => {
      try {
        return r.json("status") === "UP";
      } catch (_) {
        return false;
      }
    },
  });

  if (ok) {
    // Once healthy, we can stop early by throwing (k6 will finish the iteration)
    // In practice the outer benchmark.sh measures wall-clock time.
  }

  sleep(0.5);
}
