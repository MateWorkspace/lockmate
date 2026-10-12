#!/bin/sh
set -eu

urlencode() {
    jq -nr --arg value "$1" '$value | @uri'
}

postgres_host="${LOCKMATE_POSTGRES_HOST:-127.0.0.1}"
case "$postgres_host" in
    \[*\]) ;;
    *:*) postgres_host="[$postgres_host]" ;;
esac
postgres_password_raw="${LOCKMATE_POSTGRES_PASSWORD:-postgres}"
postgres_username="$(urlencode "${LOCKMATE_POSTGRES_USERNAME:-postgres}")"
postgres_password="$(urlencode "$postgres_password_raw")"
postgres_database="$(urlencode "${LOCKMATE_POSTGRES_DATABASE:-lockmate}")"
postgres_ssl_mode="$(urlencode "${LOCKMATE_POSTGRES_SSL_MODE:-disable}")"
postgres_timeout="${LOCKMATE_POSTGRES_CONNECT_TIMEOUT:-10s}"
postgres_url="postgres://${postgres_username}:${postgres_password}@${postgres_host}:${LOCKMATE_POSTGRES_PORT:-5432}/${postgres_database}?sslmode=${postgres_ssl_mode}"

# lib/pq accepts integer seconds; round valid Go-style durations upward.
postgres_timeout="$(jq -nr --arg value "$postgres_timeout" '
    ($value | ltrimstr("+")) as $duration |
    (if ($duration | test("^[0-9]+$")) then ($duration | tonumber)
     elif ($duration | test("^([0-9]+(\\.[0-9]+)?(ns|us|µs|μs|ms|s|m|h))+$")) then
        [$duration | scan("([0-9]+(?:\\.[0-9]+)?)(ns|us|µs|μs|ms|s|m|h)") |
            (.[0] | tonumber) *
            ({ns:0.000000001,us:0.000001,"µs":0.000001,"μs":0.000001,ms:0.001,s:1,m:60,h:3600}[.[1]])] | add
     else 10 end) |
    if . >= 1 and . <= 9223372036 then ceil else 10 end')"
postgres_url="${postgres_url}&connect_timeout=${postgres_timeout}"

umask 077
migration_output="$(mktemp)"
trap 'rm -f "$migration_output"' EXIT

if migrate -path /app/migrations -database "$postgres_url" up >"$migration_output" 2>&1; then
    status=0
else
    status=$?
fi

# Replace literal values, not regular expressions. Driver parse errors can include
# a credential-bearing URL; never replay captured output without redaction.
jq -Rrs \
    --arg uri "$postgres_url" \
    --arg password "$postgres_password_raw" \
    --arg encoded "$postgres_password" \
    'reduce [$uri, $password, $encoded][] as $secret (.;
        if $secret == "" then . else split($secret) | join("[REDACTED]") end)' \
    "$migration_output"
rm -f "$migration_output"
trap - EXIT

if [ "$status" -ne 0 ]; then
    printf '%s\n' 'Database migration failed; Lockmate will not start.' >&2
    exit "$status"
fi

exec /app/lockmate "$@"
