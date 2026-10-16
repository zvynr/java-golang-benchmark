/**
 * W5 — Realistic Helpdesk
 * 100 VUs | 60s warm-up | 5m measurement | traffic mix
 * 40% Detail | 30% List | 20% Create | 10% Statistics
 */
import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Rate, Counter } from "k6/metrics";
import { SharedArray } from "k6/data";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const VU = parseInt(__ENV.VU || "100", 10);

const latency = new Trend("helpdesk_latency", true);
const errorRate = new Rate("helpdesk_errors");
const requests = new Counter("helpdesk_requests");

// Pre-generate some ticket IDs that exist in seed (1..500)
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
//  stages: [
//      { duration: "5s", target: VU },  // warm-up
//      { duration: "1m", target: VU },   // measurement
//      { duration: "5s", target: 0 },   // ramp-down
//    ],
  thresholds: {
    helpdesk_errors: ["rate<0.02"],
    http_req_duration: ["p(95)<800", "p(99)<2000"],
  },
};

function randomTicketId() {
  return ticketIds[Math.floor(Math.random() * ticketIds.length)];
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
    title: `Load test ticket ${Date.now()}`,
    description: "Created by k6 helpdesk scenario",
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

  if (roll < 0.4) {
    res = detail();
  } else if (roll < 0.7) {
    res = list();
  } else if (roll < 0.9) {
    res = create();
  } else {
    res = statistics();
  }
//  res = statistics();

  latency.add(res.timings.duration);
  errorRate.add(res.status >= 400);
  requests.add(1);

  // Small think time to make it more realistic
  sleep(0.05 + Math.random() * 0.1);
}
