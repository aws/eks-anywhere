#!/usr/bin/env bash
# Functional check for the Harbor curated package. Runs in an in-cluster Job using the EKS-A
# diagnostic collector image (bash, curl, tar, sha256sum), so it needs no extra image.
#
# It checks the portal (HTTP 200 on /), API health, admin login, and then pushes (optional) and
# pulls back a one-layer OCI image through the registry API. Harbor's certificate is self-signed,
# so TLS verification is skipped (-k).
set -euo pipefail

: "${HARBOR_URL:?}" "${HARBOR_USERNAME:?}" "${HARBOR_PASSWORD:?}"
: "${HARBOR_REPOSITORY:?}" "${HARBOR_TAG:?}" "${HARBOR_IMAGE_CONTENT:?}"
PUSH="${HARBOR_PUSH:-true}"
TIMEOUT="${CHECK_TIMEOUT_SECONDS:-600}"
WORK="$(mktemp -d)"
CURL=(curl -sS -k --connect-timeout 10 --max-time 60)
AUTH=(-u "${HARBOR_USERNAME}:${HARBOR_PASSWORD}")
REG_AUTH=()

log() { echo "[harbor-check] $*"; }
fail() {
  echo "[harbor-check] FAIL: $*" >&2
  exit 1
}

# status URL [curl args...] prints the HTTP status code, 000 if the request failed.
status() {
  local url=$1
  shift
  "${CURL[@]}" -o /dev/null -w '%{http_code}' "$@" "$url" || true
}

health() { "${CURL[@]}" "$HARBOR_URL/api/v2.0/health" 2>&1 || true; }

healthy() {
  local h
  h=$(health)
  grep -q '"healthy"' <<<"$h" && ! grep -q 'unhealthy' <<<"$h"
}

wait_ready() {
  local deadline=$((SECONDS + TIMEOUT))
  until [[ "$(status "$HARBOR_URL/")" == 200 ]] && healthy; do
    if ((SECONDS >= deadline)); then
      fail "not ready after ${TIMEOUT}s: portal HTTP $(status "$HARBOR_URL/"), health: $(health)"
    fi
    sleep 10
  done
  log "portal returned HTTP 200 and API health is healthy: $(health)"
}

login() {
  local code
  code=$(status "$HARBOR_URL/api/v2.0/users/current" "${AUTH[@]}")
  [[ $code == 200 ]] || fail "login as ${HARBOR_USERNAME} returned HTTP $code"
  log "logged in as ${HARBOR_USERNAME}"
}

# registry_auth uses basic auth when the registry accepts it, and a bearer token otherwise.
registry_auth() {
  if [[ "$(status "$HARBOR_URL/v2/" "${AUTH[@]}")" == 200 ]]; then
    REG_AUTH=("${AUTH[@]}")
    log "registry accepts basic auth"
    return
  fi
  local token
  token=$("${CURL[@]}" "${AUTH[@]}" \
    "$HARBOR_URL/service/token?service=harbor-registry&scope=repository:${HARBOR_REPOSITORY}:pull,push" |
    sed -n 's/.*"token": *"\([^"]*\)".*/\1/p')
  [[ -n "$token" ]] || fail "could not get a registry token"
  REG_AUTH=(-H "Authorization: Bearer ${token}")
  log "using a registry bearer token"
}

sha() { sha256sum "$1" | cut -d' ' -f1; }
size() { wc -c <"$1" | tr -d ' '; }

upload_blob() {
  local file=$1 digest=$2 code loc
  code=$(status "$HARBOR_URL/v2/$HARBOR_REPOSITORY/blobs/$digest" -I "${REG_AUTH[@]}")
  if [[ $code == 200 ]]; then
    log "blob $digest already exists"
    return
  fi
  loc=$("${CURL[@]}" "${REG_AUTH[@]}" -X POST -D - -o /dev/null "$HARBOR_URL/v2/$HARBOR_REPOSITORY/blobs/uploads/" |
    tr -d '\r' | sed -n 's/^[Ll]ocation: *//p')
  [[ -n "$loc" ]] || fail "starting the upload of blob $digest returned no Location header"
  # Keep only the path and query, the Location host may be Harbor's external URL.
  case "$loc" in http://* | https://*) loc="/${loc#*://*/}" ;; esac
  case "$loc" in *\?*) loc="${loc}&" ;; *) loc="${loc}?" ;; esac
  code=$(status "${HARBOR_URL}${loc}digest=${digest/:/%3A}" "${REG_AUTH[@]}" -X PUT \
    -H 'Content-Type: application/octet-stream' --data-binary "@$file")
  [[ $code == 201 ]] || fail "uploading blob $digest returned HTTP $code"
  log "uploaded blob $digest"
}

push_image() {
  mkdir -p "$WORK/rootfs"
  printf '%s\n' "$HARBOR_IMAGE_CONTENT" >"$WORK/rootfs/eksa-e2e.txt"
  tar -C "$WORK/rootfs" -cf "$WORK/layer.tar" eksa-e2e.txt
  local layer_digest layer_size config_digest config_size code
  layer_digest="sha256:$(sha "$WORK/layer.tar")"
  layer_size=$(size "$WORK/layer.tar")
  printf '{"architecture":"amd64","os":"linux","config":{},"rootfs":{"type":"layers","diff_ids":["%s"]}}' \
    "$layer_digest" >"$WORK/config.json"
  config_digest="sha256:$(sha "$WORK/config.json")"
  config_size=$(size "$WORK/config.json")
  printf '{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json","config":{"mediaType":"application/vnd.oci.image.config.v1+json","digest":"%s","size":%s},"layers":[{"mediaType":"application/vnd.oci.image.layer.v1.tar","digest":"%s","size":%s}]}' \
    "$config_digest" "$config_size" "$layer_digest" "$layer_size" >"$WORK/manifest.json"

  upload_blob "$WORK/layer.tar" "$layer_digest"
  upload_blob "$WORK/config.json" "$config_digest"
  code=$(status "$HARBOR_URL/v2/$HARBOR_REPOSITORY/manifests/$HARBOR_TAG" "${REG_AUTH[@]}" -X PUT \
    -H 'Content-Type: application/vnd.oci.image.manifest.v1+json' --data-binary "@$WORK/manifest.json")
  [[ $code == 201 ]] || fail "pushing manifest ${HARBOR_REPOSITORY}:${HARBOR_TAG} returned HTTP $code"
  log "pushed ${HARBOR_REPOSITORY}:${HARBOR_TAG} (manifest sha256:$(sha "$WORK/manifest.json"))"
}

pull_image() {
  "${CURL[@]}" --fail "${REG_AUTH[@]}" -H 'Accept: application/vnd.oci.image.manifest.v1+json' \
    -o "$WORK/pulled-manifest.json" "$HARBOR_URL/v2/$HARBOR_REPOSITORY/manifests/$HARBOR_TAG" ||
    fail "pulling manifest ${HARBOR_REPOSITORY}:${HARBOR_TAG}"
  if [[ -f "$WORK/manifest.json" && "$(sha "$WORK/manifest.json")" != "$(sha "$WORK/pulled-manifest.json")" ]]; then
    fail "pulled manifest differs from the pushed one"
  fi
  # The manifest lists the config digest first and the layer digest second.
  local layer_digest content
  layer_digest=$(grep -o '"digest":"sha256:[0-9a-f]*"' "$WORK/pulled-manifest.json" | sed -n '2s/.*"\(sha256:[0-9a-f]*\)"/\1/p')
  [[ -n "$layer_digest" ]] || fail "no layer in pulled manifest: $(cat "$WORK/pulled-manifest.json")"
  "${CURL[@]}" --fail -L "${REG_AUTH[@]}" -o "$WORK/pulled-layer.tar" \
    "$HARBOR_URL/v2/$HARBOR_REPOSITORY/blobs/$layer_digest" || fail "pulling layer $layer_digest"
  [[ "sha256:$(sha "$WORK/pulled-layer.tar")" == "$layer_digest" ]] || fail "pulled layer does not match digest $layer_digest"
  content=$(tar -xOf "$WORK/pulled-layer.tar" eksa-e2e.txt)
  [[ "$content" == "$HARBOR_IMAGE_CONTENT" ]] || fail "pulled image content is '$content', expected '$HARBOR_IMAGE_CONTENT'"
  log "pulled ${HARBOR_REPOSITORY}:${HARBOR_TAG} and verified its content"
}

wait_ready
login
registry_auth
if [[ "$PUSH" == "true" ]]; then
  push_image
fi
pull_image
log "PASS"
