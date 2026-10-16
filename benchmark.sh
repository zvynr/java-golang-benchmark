#!/usr/bin/env bash
set -euo pipefail

# Disable Git Bash / MSYS2 path conversion on Windows when passing paths to Docker
export MSYS_NO_PATHCONV=1
export MSYS2_ARG_CONV_EXCL="*"

# ═══════════════════════════════════════════════════════════════════════════════
#  Java 21 JVM vs GraalVM Native 21 vs Go — Standalone Benchmark Runner
# ═══════════════════════════════════════════════════════════════════════════════

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_DIR="results"
TIMESTAMP="$(date +%Y%m%d_%H%M%S)"
RUN_DIR="${RESULTS_DIR}/${TIMESTAMP}"

# Defaults
RUNTIME="${RUNTIME:-both}"          # jvm | native | go | both | all
WORKLOAD="${1:-all}"                # all | W0 | W1 | W2 | W3 | W4 | W5 | W6 | W7
REPEATS="${REPEATS:-1}"
K6_BIN="${K6_BIN:-k6}"

JVM_URL="http://localhost:8080"
NATIVE_URL="http://localhost:8081"
GO_URL="http://localhost:8082"

mkdir -p "${RUN_DIR}"

log()  { echo -e "\033[1;34m[INFO]\033[0m  $*"; }
ok()   { echo -e "\033[1;32m[OK]\033[0m    $*"; }
warn() { echo -e "\033[1;33m[WARN]\033[0m  $*"; }
err()  { echo -e "\033[1;31m[ERR]\033[0m   $*"; }

# ── Helpers ───────────────────────────────────────────────────────────────────

wait_healthy() {
  local url="$1"
  local name="$2"
  local max_wait="${3:-120}"
  local elapsed=0

  log "Waiting for ${name} to become healthy (${url}/actuator/health)..."
  while [[ $elapsed -lt $max_wait ]]; do
    if curl -sf "${url}/actuator/health" 2>/dev/null | grep -q '"status":"UP"'; then
      ok "${name} is UP (took ~${elapsed}s)"
      echo "${elapsed}" > "${RUN_DIR}/${name}_time_to_ready.txt"
      return 0
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  err "${name} did not become healthy within ${max_wait}s"
  return 1
}

start_stack() {
  local profile="$1"   # jvm | native | go
  log "Starting postgres + app-${profile}..."
  docker compose --profile "${profile}" up -d --build postgres "app-${profile}"
}

stop_app() {
  local profile="$1"
  log "Stopping app-${profile}..."
  docker compose --profile "${profile}" stop "app-${profile}" 2>/dev/null || true
  docker compose --profile "${profile}" rm -f "app-${profile}" 2>/dev/null || true
  docker stop "benchmark-${profile}" 2>/dev/null || true
  docker rm -f "benchmark-${profile}" 2>/dev/null || true
}

stop_all() {
  log "Stopping all services..."
  docker compose --profile jvm --profile native --profile go down 2>/dev/null || true
}

run_k6() {
  local script="$1"
  local base_url="$2"
  local runtime="$3"
  local workload="$4"
  local vu="${5:-}"
  local extra_env="${6:-}"

  local out_dir="${RUN_DIR}/${runtime}/${workload}"
  mkdir -p "${out_dir}"

  local vu_flag=""
  if [[ -n "$vu" ]]; then
    vu_flag="VU=${vu}"
    out_dir="${out_dir}/vu${vu}"
    mkdir -p "${out_dir}"
  fi

  log "k6 ${workload} | runtime=${runtime} | vu=${vu:--} | url=${base_url}"

  if command -v "${K6_BIN}" >/dev/null 2>&1; then
    # shellcheck disable=SC2086
    BASE_URL="${base_url}" ${vu_flag} ${extra_env} \
      "${K6_BIN}" run \
        --summary-export="${out_dir}/summary.json" \
        "k6/${script}" \
      | tee "${out_dir}/k6_stdout.txt" || true
  else
    warn "'${K6_BIN}' not found on host. Falling back to Docker Compose k6 container..."
    docker compose --profile k6 run --rm \
      -e BASE_URL="${base_url}" \
      ${vu_flag:+ -e VU="${vu}"} \
      k6 run --summary-export="/results/${TIMESTAMP}/${runtime}/${workload}/summary.json" "/scripts/${script}" \
      | tee "${out_dir}/k6_stdout.txt" || true
  fi
}

run_workload() {
  local wl="$1"
  local runtime="$2"
  local url="$3"

  case "${wl}" in
    W0)
      # Cold start is measured during startup
      ok "W0 Cold Start measured during container startup."
      ;;
    W1)
      run_k6 "w1-cpu.js" "${url}" "${runtime}" "W1"
      ;;
    W2)
      run_k6 "w2-io.js" "${url}" "${runtime}" "W2"
      ;;
    W3)
      run_k6 "w3-concurrency.js" "${url}" "${runtime}" "W3" "50"
      run_k6 "w3-concurrency.js" "${url}" "${runtime}" "W3" "100"
      run_k6 "w3-concurrency.js" "${url}" "${runtime}" "W3" "200"
      ;;
    W4)
      run_k6 "w4-allocation.js" "${url}" "${runtime}" "W4"
      ;;
    W5)
      run_k6 "w5-helpdesk.js" "${url}" "${runtime}" "W5"
      ;;
    W5_1)
      run_k6 "w7-helpdesk.js" "${url}" "${runtime}" "W5_1"
      ;;
    W6)
      run_k6 "w6-hybrid.js" "${url}" "${runtime}" "W6"
      ;;
    *)
      err "Unknown workload: ${wl}"
      exit 1
      ;;
  esac
}

get_runtimes() {
  case "${RUNTIME}" in
    jvm)       echo "jvm" ;;
    native)    echo "native" ;;
    go)        echo "go" ;;
    both)      echo "jvm native" ;;
    all)       echo "jvm native go" ;;
    *)         err "Unknown RUNTIME: ${RUNTIME}"; exit 1 ;;
  esac
}

# ── Main ──────────────────────────────────────────────────────────────────────

usage() {
  cat <<EOF
Usage: $0 [WORKLOAD]

  WORKLOAD   all (default) | W0 | W1 | W2 | W3 | W4 | W5 | W5_1 | W6 | W7

Environment variables:
  RUNTIME=jvm|native|go|both|all   (default: both)
  REPEATS=N                        (default: 1)
  K6_BIN=path/to/k6

Examples:
  ./benchmark.sh all
  RUNTIME=jvm ./benchmark.sh W1
  RUNTIME=native REPEATS=3 ./benchmark.sh W3
  RUNTIME=go ./benchmark.sh W5
EOF
}

main() {
  if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    usage
    exit 0
  fi

  local total_start_ms
  total_start_ms=$(date +%s%3N 2>/dev/null || echo $(($(date +%s) * 1000)))

  log "Results will be written to: ${RUN_DIR}"
  log "Runtime=${RUNTIME}  Workload=${WORKLOAD}  Repeats=${REPEATS}"

  # Ensure postgres & network are up first
  docker network create benchmark-net 2>/dev/null || true
  docker volume create pgdata 2>/dev/null || true
  docker compose up -d postgres
  sleep 5

  for wl in W1 W2 W3 W4 W5 W5_1 W6; do
    if [[ "${WORKLOAD}" == "all" || "${WORKLOAD}" == "${wl}" ]]; then
      log "═══════════════════════════════════════════════════════════════"
      log "  Workload: ${wl}"
      log "═══════════════════════════════════════════════════════════════"

      local wl_loop_start_ms
      wl_loop_start_ms=$(date +%s%3N 2>/dev/null || echo $(($(date +%s) * 1000)))

      log "Resetting & starting monitoring stack for workload ${wl}..."
      docker compose --profile monitoring down -v 2>/dev/null || true
      docker compose --profile monitoring up -d 2>/dev/null || true
      sleep 5

      for runtime in $(get_runtimes); do
        local url=""
        case "${runtime}" in
          jvm)      url="${JVM_URL}" ;;
          native)   url="${NATIVE_URL}" ;;
          go)       url="${GO_URL}" ;;
        esac

        log "████████████████████████████████████████████████████████████"
        log "  Workload: ${wl}  |  Runtime: ${runtime}  |  URL: ${url}"
        log "████████████████████████████████████████████████████████████"

        start_sdk_or_stack() {
          start_stack "${runtime}"
        }
        start_stack "${runtime}"
        wait_healthy "${url}" "${runtime}" 180

        local wl_start_ms
        wl_start_ms=$(date +%s%3N 2>/dev/null || echo $(($(date +%3N) * 1000))) 2>/dev/null || wl_start_ms=$(date +%s)000

        for ((i=1; i<=REPEATS; i++)); do
          log "── ${wl} run ${i}/${REPEATS} (${runtime}) ──"
          run_workload "${wl}" "${runtime}" "${url}"
        done

        stop_app "${runtime}"
      done

      docker compose --profile monitoring down -v 2>/dev/null || true
    fi
  done

  stop_all
  ok "All benchmarks completed successfully! Results in ${RUN_DIR}"
}

main "$@"
