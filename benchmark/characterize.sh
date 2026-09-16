#!/usr/bin/env bash
# Systematic characterization of redis-go vs real Redis.
# Sweeps across concurrency levels and payload sizes, producing a CSV
# that can be imported into any analysis tool.
#
# Usage:
#   bash benchmark/characterize.sh
#   REQUESTS=5000 bash benchmark/characterize.sh   # faster pass
#
# Output:
#   benchmark/results/<ts>/char_results.csv
#   benchmark/results/<ts>/char_<server>_<sweep>_<cmd>_c<N>_p<B>.txt  (raw)

set -uo pipefail

NETWORK="redis-char-net"
REDIS_OFFICIAL_CTR="char-redis-official"
REDIS_GO_CTR="char-redis-go"

REQUESTS="${REQUESTS:-10000}"
WARMUP=1000

# ── Sweep definitions ──────────────────────────────────────────────────────────
# Concurrency sweep: vary clients, fix payload at 32B, commands: set, get
CONC_CMDS=(set get)
CONC_CLIENTS=(1 10 50 100)
CONC_PAYLOAD=32

# Payload sweep: vary payload, fix clients at 50, commands: set, get
PAY_CMDS=(set get)
PAY_SIZES=(32 256 4096)
PAY_CLIENTS=50

# ── Output ─────────────────────────────────────────────────────────────────────
TS=$(date +%Y%m%d_%H%M%S)
OUT="benchmark/results/$TS"
mkdir -p "$OUT"
CSV="$OUT/char_results.csv"
echo "server,sweep,command,clients,payload_bytes,throughput,avg_ms,p50_ms,p95_ms,p99_ms,cpu_avg_pct,cpu_peak_pct,mem_peak_mib" > "$CSV"

# ── Colours ────────────────────────────────────────────────────────────────────
CYAN='\033[0;36m'; YELLOW='\033[1;33m'; GREEN='\033[0;32m'; NC='\033[0m'
log()  { echo -e "${CYAN}[char]${NC} $*"; }
step() { echo -e "\n${YELLOW}▶ $*${NC}"; }

# ── Cleanup ────────────────────────────────────────────────────────────────────
cleanup() {
    log "Stopping containers..."
    docker rm -f "$REDIS_OFFICIAL_CTR" "$REDIS_GO_CTR" 2>/dev/null || true
    docker network rm "$NETWORK" 2>/dev/null || true
}
trap cleanup EXIT

# ── Start servers ──────────────────────────────────────────────────────────────
log "Creating network $NETWORK"
docker network create "$NETWORK" 2>/dev/null || true

log "Starting redis-official (no persistence)..."
docker run -d --name "$REDIS_OFFICIAL_CTR" --network "$NETWORK" \
    --cpus="1" --memory="256m" \
    redis:7-alpine redis-server --save "" --appendonly no > /dev/null

log "Building and starting redis-go..."
docker build -t redis-go-bench "$(dirname "$0")/.." -q
docker run -d --name "$REDIS_GO_CTR" --network "$NETWORK" \
    --cpus="1" --memory="256m" \
    redis-go-bench > /dev/null

wait_for() {
    printf "  Waiting for %-30s" "$1..."
    for _ in $(seq 1 40); do
        docker run --rm --network "$NETWORK" redis:7-alpine \
            redis-cli -h "$1" PING 2>/dev/null | grep -q PONG && { echo -e " ${GREEN}ok${NC}"; return 0; }
        sleep 0.5; printf "."
    done
    echo " TIMEOUT" >&2; return 1
}
wait_for "$REDIS_OFFICIAL_CTR"
wait_for "$REDIS_GO_CTR"

# ── Core measurement function ──────────────────────────────────────────────────
# run_one <server_ctr> <label> <sweep> <cmd> <clients> <payload>
# Appends one CSV row and saves raw output.
run_one() {
    local ctr=$1 label=$2 sweep=$3 cmd=$4 clients=$5 payload=$6
    local tag="${label}_${sweep}_${cmd}_c${clients}_p${payload}"
    local raw="$OUT/char_${tag}.txt"
    local stats_f="$OUT/char_${tag}_stats.txt"

    # Warmup (discard)
    docker run --rm --network "$NETWORK" redis:7-alpine \
        redis-benchmark -h "$ctr" -n "$WARMUP" -c "$clients" -d "$payload" \
        -t "$cmd" -q > /dev/null 2>&1 || true

    # Background stats
    > "$stats_f"
    while true; do
        docker stats --no-stream \
            --format "{{.CPUPerc}} {{.MemUsage}}" \
            "$ctr" 2>/dev/null >> "$stats_f" || break
        sleep 0.5
    done &
    local stats_pid=$!

    # Benchmark
    docker run --rm --network "$NETWORK" redis:7-alpine \
        redis-benchmark \
            -h "$ctr" \
            -n "$REQUESTS" \
            -c "$clients" \
            -d "$payload" \
            -t "$cmd" \
        > "$raw" 2>&1 || true

    kill "$stats_pid" 2>/dev/null || true
    wait "$stats_pid" 2>/dev/null || true

    # Parse throughput and latency from raw output
    local tput avg p50 p95 p99
    tput=$(awk '/throughput summary:/{print $3; exit}' "$raw")
    read -r avg _ p50 p95 p99 _ < <(
        awk '/latency summary/{found=1} found && /^[[:space:]]+[0-9]/{print; exit}' "$raw"
    )

    # Parse resource usage
    local cpu_avg cpu_pk mem_pk
    read -r cpu_avg cpu_pk mem_pk < <(
        awk '{
            gsub(/%/,"",$1); c=$1+0; cs+=c; n++; if(c>cp) cp=c
            m=$2; sub(/MiB.*/,"",m)
            if($2~/GiB/){gsub(/GiB.*/,"",m); m=m*1024}
            if(m+0>mp) mp=m+0
        } END { printf "%.1f %.1f %.1f\n", cs/n, cp, mp }' "$stats_f"
    )

    printf "    %-12s c=%-4s p=%-5s → %6.0f req/s  p50=%5.1fms p95=%6.1fms p99=%6.1fms  CPU avg=%.0f%%\n" \
        "$cmd" "$clients" "${payload}B" "${tput:-0}" "${p50:-0}" "${p95:-0}" "${p99:-0}" "${cpu_avg:-0}"

    echo "${label},${sweep},${cmd},${clients},${payload},${tput:-},${avg:-},${p50:-},${p95:-},${p99:-},${cpu_avg:-},${cpu_pk:-},${mem_pk:-}" >> "$CSV"
}

# ── Sweep A: Concurrency ───────────────────────────────────────────────────────
step "Sweep A — Concurrency  (${REQUESTS} req, ${CONC_PAYLOAD}B payload)"
for label in redis-official redis-go; do
    [[ $label == "redis-official" ]] && ctr=$REDIS_OFFICIAL_CTR || ctr=$REDIS_GO_CTR
    echo "  [$label]"
    for cmd in "${CONC_CMDS[@]}"; do
        for clients in "${CONC_CLIENTS[@]}"; do
            run_one "$ctr" "$label" "concurrency" "$cmd" "$clients" "$CONC_PAYLOAD"
        done
    done
done

# ── Sweep B: Payload size ──────────────────────────────────────────────────────
step "Sweep B — Payload size  (${REQUESTS} req, ${PAY_CLIENTS} clients)"
for label in redis-official redis-go; do
    [[ $label == "redis-official" ]] && ctr=$REDIS_OFFICIAL_CTR || ctr=$REDIS_GO_CTR
    echo "  [$label]"
    for cmd in "${PAY_CMDS[@]}"; do
        for payload in "${PAY_SIZES[@]}"; do
            run_one "$ctr" "$label" "payload" "$cmd" "$PAY_CLIENTS" "$payload"
        done
    done
done

# ── Summary table ──────────────────────────────────────────────────────────────
echo ""
echo "================================================================"
echo "  CHARACTERIZATION COMPLETE"
echo "  CSV: $CSV"
echo "================================================================"

# Concurrency table
echo ""
printf "%-18s %-8s %-5s │ %8s %7s %7s │ %8s %7s %7s\n" \
    "CONCURRENCY" "CMD" "C" "go tput" "p50" "p99" "off tput" "p50" "p99"
printf "%s\n" "──────────────────────────────────────────────────────────────────"
awk -F',' '
NR>1 && $2=="concurrency" {
    key=$3","$4
    if($1~/go/)  { go_t[key]=$6; go_p50[key]=$8;  go_p99[key]=$10 }
    else         { of_t[key]=$6; of_p50[key]=$8;  of_p99[key]=$10 }
}
END {
    for(k in go_t) {
        split(k,a,",")
        printf "%-18s %-8s %-5s │ %8.0f %7.1f %7.1f │ %8.0f %7.1f %7.1f\n",
            "", a[1], a[2], go_t[k], go_p50[k], go_p99[k], of_t[k], of_p50[k], of_p99[k]
    }
}' "$CSV" | sort

# Payload table
echo ""
printf "%-18s %-8s %-7s │ %8s %7s %7s │ %8s %7s %7s\n" \
    "PAYLOAD" "CMD" "BYTES" "go tput" "p50" "p99" "off tput" "p50" "p99"
printf "%s\n" "──────────────────────────────────────────────────────────────────"
awk -F',' '
NR>1 && $2=="payload" {
    key=$3","$5
    if($1~/go/)  { go_t[key]=$6; go_p50[key]=$8;  go_p99[key]=$10 }
    else         { of_t[key]=$6; of_p50[key]=$8;  of_p99[key]=$10 }
}
END {
    for(k in go_t) {
        split(k,a,",")
        printf "%-18s %-8s %-7s │ %8.0f %7.1f %7.1f │ %8.0f %7.1f %7.1f\n",
            "", a[1], a[2], go_t[k], go_p50[k], go_p99[k], of_t[k], of_p50[k], of_p99[k]
    }
}' "$CSV" | sort

echo ""
echo "Run 'make bench-compare' for the full command-type baseline."
