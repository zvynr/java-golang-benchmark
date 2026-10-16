/**
 * W6 — Full Hybrid Workload
 * 60s warm-up | 5m measurement | 30s ramp-down
 * Traffic Mix:
 *   20% CPU Bound        (POST /api/benchmark/cpu)
 *   20% Allocation Bound (POST /api/benchmark/allocation)
 *   15% Ticket Detail    (GET /api/tickets/{id})
 *   15% Ticket List      (GET /api/tickets?page=0&size=20)
 *   15% Ticket Create    (POST /api/tickets)
 *   15% Ticket Statistics(GET /api/tickets/statistics)
 */
import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Rate, Counter } from "k6/metrics";
import { SharedArray } from "k6/data";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const VU = parseInt(__ENV.VU || "100", 10);

const latency = new Trend("hybrid_latency", true);
const errorRate = new Rate("hybrid_errors");
const requests = new Counter("hybrid_requests");

// Pre-generate ticket IDs (1..500)
const ticketIds = new SharedArray("ticketIds", function () {
  const ids = [];
  for (let i = 1; i <= 500; i++) ids.push(i);
  return ids;
});

export const options = {
  stages: [
      { duration: "60s", target: VU },  // warm-up
      { duration: "5m", target: VU },   // measurement
      { duration: "30s", target: 0 },   // ramp-down
    ],
  thresholds: {
    hybrid_errors: ["rate<0.02"],
    http_req_duration: ["p(95)<1500", "p(99)<3000"],
  },
};

function randomTicketId() {
  return ticketIds[Math.floor(Math.random() * ticketIds.length)];
}

function cpu() {
  const payload = JSON.stringify({
    iterations: 5000,
    hashRounds: 3,
    sortSize: 2500,
  });
  const res = http.post(`${BASE_URL}/api/benchmark/cpu`, payload, {
    headers: { "Content-Type": "application/json" },
    tags: { name: "cpu", endpoint: "POST /benchmark/cpu" },
  });
  check(res, { "cpu 200": (r) => r.status === 200 });
  return res;
}

function allocation() {
  const payload = JSON.stringify({
    objectCount: 5000,
    objectSizeBytes: 256,
    transformRounds: 2,
  });
  const res = http.post(`${BASE_URL}/api/benchmark/allocation`, payload, {
    headers: { "Content-Type": "application/json" },
    tags: { name: "allocation", endpoint: "POST /benchmark/allocation" },
  });
  check(res, { "allocation 200": (r) => r.status === 200 });
  return res;
}

function detail() {
  const id = randomTicketId();
  const res = http.get(`${BASE_URL}/api/tickets/${id}`, {
    tags: { name: "detail", endpoint: "GET /tickets/{id}" },
  });
  check(res, { "detail 200 or 404": (r) => r.status === 200 || r.status === 404 });
  return res;
}

function list() {
  const page = Math.floor(Math.random() * 5);
  const res = http.get(`${BASE_URL}/api/tickets?page=${page}&size=20`, {
    tags: { name: "list", endpoint: "GET /tickets" },
  });
  check(res, { "list 200": (r) => r.status === 200 });
  return res;
}

function create() {
  const priorities = ["LOW", "MEDIUM", "HIGH", "CRITICAL"];
  const priority = priorities[Math.floor(Math.random() * priorities.length)];
  const payload = JSON.stringify({
    title: `Hybrid test ticket ${Date.now()}`,
    description: "Created by k6 w6 hybrid scenario",
    priority: priority,
    assignee: `user${Math.floor(Math.random() * 10) + 1}`,
  });
  const res = http.post(`${BASE_URL}/api/tickets`, payload, {
    headers: { "Content-Type": "application/json" },
    tags: { name: "create", endpoint: "POST /tickets" },
  });
  check(res, { "create 201": (r) => r.status === 201 });
  return res;
}

function statistics() {
  const res = http.get(`${BASE_URL}/api/tickets/statistics`, {
    tags: { name: "statistics", endpoint: "GET /tickets/statistics" },
  });
  check(res, { "stats 200": (r) => r.status === 200 });
  return res;
}

export default function () {
  const roll = Math.random();
  let res;

  if (roll < 0.05) {
    // 20% CPU Bound
    res = cpu();
  } else if (roll < 0.10) {
    // 20% Allocation Bound
    res = allocation();
  } else if (roll < 0.40) {
    // 15% Ticket Detail
    res = detail();
  } else if (roll < 0.70) {
    // 15% Ticket List
    res = list();
  } else if (roll < 0.90) {
    // 15% Ticket Create
    res = create();
  } else {
    // 15% Ticket Statistics
    res = statistics();
  }

  latency.add(res.timings.duration);
  errorRate.add(res.status >= 400);
  requests.add(1);

  // Think time
  sleep(0.05 + Math.random() * 0.1);
}
