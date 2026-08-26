#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
workspace_dir=$(CDPATH= cd -- "$script_dir/../.." && pwd)
compose_dir="$workspace_dir/deploy/compose"
compose_env=${GO_ADMIN_COMPOSE_ENV:-"$compose_dir/.env"}
base_url=${GO_ADMIN_BASE_URL:-"http://127.0.0.1:${GO_ADMIN_WEB_PORT:-8080}"}

compose() {
  docker compose --env-file "$compose_env" \
    -f "$compose_dir/compose.yml" \
    -f "$compose_dir/compose.build.yml" "$@"
}

wait_status() {
  url=$1
  expected=$2
  attempts=${3:-40}
  count=0
  while [ "$count" -lt "$attempts" ]; do
    status=$(curl -sS -o /dev/null -w '%{http_code}' "$url" || true)
    if [ "$status" = "$expected" ]; then
      return 0
    fi
    count=$((count + 1))
    sleep 2
  done
  echo "$url returned $status, expected $expected" >&2
  return 1
}

wait_status "$base_url/health/live" 200
wait_status "$base_url/health/ready" 200
echo "[verify] health endpoints ready"

capabilities=$(curl -fsS "$base_url/api/v1/runtime/capabilities")
echo "$capabilities" | jq -c '{hostProfile,desktop,offline}'
echo "$capabilities" | jq -e '.hostProfile == "server" and .desktop == false and .offline == false' >/dev/null
echo "[verify] server capabilities"

captcha=$(curl -fsS "$base_url/api/v1/captcha")
captcha_id=$(echo "$captcha" | jq -er 'select(.code == 200) | .id | select(length > 0)')
captcha_answer=$(compose exec -T redis redis-cli --raw GET "$captcha_id")
test -n "$captcha_answer"
login_payload=$(jq -cn --arg code "$captcha_answer" --arg uuid "$captcha_id" \
  '{username:"admin",password:"123456",code:$code,uuid:$uuid}')
login=$(curl -fsS -H 'Content-Type: application/json' \
  --data "$login_payload" \
  "$base_url/api/v1/login")
unset captcha_answer login_payload
echo "$login" | jq -c '{code,has_token:((.token // "") | length > 0)}'
token=$(echo "$login" | jq -er 'select(.code == 200) | .token | select(length > 0)')
echo "[verify] same-origin login"

release_code="COMPOSE-$(date +%s)"
created=$(curl -fsS -H 'Content-Type: application/json' -H "Authorization: Bearer $token" \
  --data "{\"name\":\"Compose persistence probe\",\"code\":\"$release_code\",\"price\":1.25,\"status\":\"2\",\"remark\":\"release verification\"}" \
  "$base_url/api/v1/demo-product")
echo "$created" | jq -c '{code,data_type:(.data | type)}'
record_id=$(echo "$created" | jq -er 'select(.code == 200) | .data | select(. > 0)')
echo "[verify] persistence fixture created"

compose restart api >/dev/null
wait_status "$base_url/health/ready" 200
detail=$(curl -fsS -H "Authorization: Bearer $token" "$base_url/api/v1/demo-product/$record_id")
echo "$detail" | jq -c '{code,persisted_code:(.data.code // null)}'
echo "$detail" | jq -e --arg code "$release_code" '.code == 200 and .data.code == $code' >/dev/null
echo "[verify] database survives API restart"

api_404=$(curl -sS -o /dev/null -w '%{http_code}' "$base_url/api/v1/route-that-does-not-exist")
test "$api_404" = 404
curl -fsS "$base_url/deep/spa/route" | grep -q 'id="app"'
echo "[verify] backend 404 and SPA fallback remain distinct"

api_container=$(compose ps -q api)
web_container=$(compose ps -q web)
test -n "$api_container"
test -n "$web_container"
for container in "$api_container" "$web_container"; do
  test "$(docker inspect -f '{{.HostConfig.Privileged}}' "$container")" = false
  test "$(docker inspect -f '{{.HostConfig.ReadonlyRootfs}}' "$container")" = true
  runtime_user=$(docker inspect -f '{{.Config.User}}' "$container")
  test -n "$runtime_user"
  test "$runtime_user" != 0
  test "$runtime_user" != root
done
echo "[verify] non-root read-only containers"

compose stop redis >/dev/null
wait_status "$base_url/health/ready" 503
compose start redis >/dev/null
wait_status "$base_url/health/ready" 200
echo "[verify] Redis readiness degradation and recovery"

echo "GO_ADMIN_LINUX_COMPOSE_E2E_PASS"
