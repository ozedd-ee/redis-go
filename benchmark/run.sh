#!/usr/bin/env bash
# Compares redis-go against real Redis (no persistence) under identical conditions.
#
# Requirements: Docker running.
# Both servers run inside a Docker bridge network. redis-benchmark runs as an
# ephemeral container on the same network so latency reflects server perf, not
# host↔container routing.
#
# Results: benchmark/results/<timestamp>/

set -uo pipefail   # note: no -e; benchmark failures are handled explicitly

NETWORK="redis-bench-net"
REDIS_OFFICIAL_CTR="bench-redis-official"
REDIS_GO_CTR="bench-redis-go"

# ── Tuning ─────────────────────────────────────────────────────────────────────
# These defaults complete in ~60-120 s even on Docker Desktop for Windows.
# Scale up BENCH_REQUESTS for more statistical precision once you know it runs.
BENCH_REQUESTS="${BENCH_REQUESTS:-10000}"
BENCH_CLIENTS="${BENCH_CLIENTS:-50}"
BENCH_PAYLOAD="${BENCH_PAYLOAD:-32}"          # bytes
WARMUP_REQUESTS=1000
# Commands to test (each runs as a separate redis-benchmark invocation so
# results appear incrementally rather than only at the very end).
BENCH_CMDS=(set get incr lpush lrange)

# ── Output directory ───────────────────────────────────────────────────────────
TS=$(date +%Y%m%d_%H%M%S)
RESULTS="benchmark/results/$TS"
mkdir -p "$RESULTS"

# ── Colours ────────────────────────────────────────────────────────────────────
BOLD='\033[1m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; CYAN='\033[0;36m'; RED='\033[0;31m'; NC='\033[0m'
log()   { echo -e "${CYAN}[bench]${NC} $*"; }
info()  { echo -e "${GREEN}  ✓${NC} $*"; }
warn()  { echo -e "${YELLOW}  !${NC} $*"; }
hdr()   { echo -e "\n${BOLD}${YELLOW}=== $* ===${NC}"; }

# ── Cleanup ────────────────────────────────────────────────────────────────────
cleanup() {
    echo ""
    log "Stopping containers..."
    docker rm -f "$REDIS_OFFICIAL_CTR" "$REDIS_GO_CTR" 2>/dev/null || true
    docker network rm "$NETWORK" 2>/dev/null || true
}
trap cleanup EXIT

# ── Setup ──────────────────────────────────────────────────────────────────────
log "Creating bridge network: $NETWORK"
docker network create "$NETWORK" 2>/dev/null || true

log "Starting real Redis (persistence disabled)..."
docker run -d \
    --name "$REDIS_OFFICIAL_CTR" \
    --network "$NETWORK" \
    --cpus="1" --memory="256m" \
    redis:7-alpine \
    redis-server --save "" --appendonly no \
    > /dev/null

log "Building and starting redis-go..."
docker build -t redis-go-bench "$(dirname "$0")/.." -q
docker run -d \
    --name "$REDIS_GO_CTR" \
    --network "$NETWORK" \
    --cpus="1" --memory="256m" \
    redis-go-bench \
    > /dev/null

# ── Readiness check ────────────────────────────────────────────────────────────
wait_for() {
    local ctr=$1
    printf "  Waiting for %-32s" "$ctr..."
    for _ in $(seq 1 40); do
        if docker run --rm --network "$NETWORK" redis:7-alpine \
               redis-cli -h "$ctr" PING 2>/dev/null | grep -q PONG; then
            echo -e " ${GREEN}ready${NC}"
            return 0
        fi
        sleep 0.5; printf "."
    done
    echo -e " ${RED}TIMEOUT${NC}" >&2; return 1
}

wait_for "$REDIS_OFFICIAL_CTR"
wait_for "$REDIS_GO_CTR"

# ── Stats collector (background) ───────────────────────────────────────────────
# Polls docker stats every ~1 s. We start/stop it around each server's benchmark.
_stats_pid=""
start_stats() {
    local ctr=$1 outfile=$2
    > "$outfile"
    while true; do
        docker stats --no-stream \
            --format "{{.CPUPerc}} {{.MemUsage}}" \
            "$ctr" 2>/dev/null >> "$outfile" || break
        sleep 0.5
    done &
    _stats_pid=$!
}
stop_stats() {
    kill "$_stats_pid" 2>/dev/null || true
    wait "$_stats_pid" 2>/dev/null || true
    _stats_pid=""
}

# ── Benchmark one server ────────────────────────────────────────────────────────
# Runs each command type individually so results appear as they complete.
bench_server() {
    local ctr=$1 label=$2
    local bench_out="$RESULTS/${label}_bench.txt"
    local stats_out="$RESULTS/${label}_stats.txt"

    hdr "$label"
    log "Params: ${BENCH_REQUESTS} reqs × ${#BENCH_CMDS[@]} cmds, ${BENCH_CLIENTS} clients, ${BENCH_PAYLOAD}B payload"

    # Warmup — ensures TCP connection pools are warm before we measure
    log "Warmup (${WARMUP_REQUESTS} reqs)..."
    if ! docker run --rm --network "$NETWORK" redis:7-alpine \
            redis-benchmark -h "$ctr" \
            -n "$WARMUP_REQUESTS" -c "$BENCH_CLIENTS" \
            -d "$BENCH_PAYLOAD" -t set,get -q > /dev/null 2>&1; then
        warn "Warmup failed for $label — skipping"
        return 1
    fi
    info "Warmup complete"

    > "$bench_out"
    start_stats "$ctr" "$stats_out"

    local cmd ok=0 fail=0
    for cmd in "${BENCH_CMDS[@]}"; do
        log "  Testing: $cmd"
        if docker run --rm --network "$NETWORK" redis:7-alpine \
               redis-benchmark \
                   -h "$ctr" \
                   -n "$BENCH_REQUESTS" \
                   -c "$BENCH_CLIENTS" \
                   -d "$BENCH_PAYLOAD" \
                   -t "$cmd" \
               2>&1 | tee -a "$bench_out"; then
            (( ok++ )) || true
        else
            warn "redis-benchmark exited non-zero for cmd '$cmd' on $label"
            (( fail++ )) || true
        fi
    done

    stop_stats

    if [[ $fail -gt 0 ]]; then
        warn "$fail/${#BENCH_CMDS[@]} command(s) failed for $label"
    else
        info "All ${#BENCH_CMDS[@]} commands completed for $label"
    fi
}

bench_server "$REDIS_OFFICIAL_CTR" "redis-official"
bench_server "$REDIS_GO_CTR"       "redis-go"

# ── Summary ────────────────────────────────────────────────────────────────────
summarise() {
    local label=$1
    local bench_out="$RESULTS/${label}_bench.txt"
    local stats_out="$RESULTS/${label}_stats.txt"

    echo ""
    echo -e "${BOLD}── $label${NC}"

    if [[ ! -s "$bench_out" ]]; then
        warn "No benchmark output recorded"
        return
    fi

    echo "  Throughput (req/sec):"
    grep "requests per second" "$bench_out" | sed 's/^/    /'

    echo "  Latency (msec) — avg / min / p50 / p95 / p99 / max:"
    if grep -q "latency summary" "$bench_out" 2>/dev/null; then
        grep -A 3 "latency summary" "$bench_out" | sed 's/^/    /'
    else
        grep "<= " "$bench_out" | tail -6 | sed 's/^/    /'
    fi

    if [[ -s "$stats_out" ]]; then
        echo "  CPU (avg / peak during benchmark):"
        awk '{
            gsub(/%/, "", $1); v=$1+0
            sum+=v; n++; if(v>peak) peak=v
        } END {
            if(n) printf "    avg=%.1f%%  peak=%.1f%%\n", sum/n, peak
        }' "$stats_out"

        echo "  Memory (peak / final):"
        awk '{
            m=$2; sub(/MiB.*/, "", m); sub(/GiB.*/, "", m)
            if ($2~/GiB/) m=m*1024
            m=m+0; if(m>pk) pk=m; last=m
        } END {
            if(pk) printf "    peak=%.1f MiB  final=%.1f MiB\n", pk, last
        }' "$stats_out"
    fi
}

echo ""
echo "================================================================"
echo -e "${BOLD}${GREEN}  SUMMARY${NC}  —  $(date)"
echo "  Requests: $BENCH_REQUESTS per command  |  Clients: $BENCH_CLIENTS  |  Payload: ${BENCH_PAYLOAD}B"
echo "================================================================"

summarise "redis-official"
summarise "redis-go"

echo ""
echo "  Raw results: $RESULTS/"
echo "================================================================"
