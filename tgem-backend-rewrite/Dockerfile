# syntax=docker/dockerfile:1.7
#
# Multi-stage build for the tgem backend (go 1.25.7, no CGO).
#
# Builder: golang:1.25-alpine. Compiles a static binary with the timetzdata
# build tag so time.LoadLocation("Asia/Dushanbe") works without /usr/share/
# zoneinfo present on the runtime image.
#
# Runtime: alpine:3.20 + a non-root user. Alpine over distroless because the
# entrypoint needs /bin/sh to template the App_dev.yaml config from env vars
# at start time (the rewrite's viper loader has no env-var override yet).
#
# Build args:
#   GO_VERSION       — pinned via the FROM tag. Update both places to bump.
#   ALPINE_VERSION   — pinned via the runtime FROM tag.

# ---------- builder ----------
FROM golang:1.25-alpine AS builder

# git is needed by `go mod download` for any VCS-resolved deps; ca-certificates
# is needed for HTTPS dep fetches. Both end up unused in the runtime image.
RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Module cache layer: only invalidated when go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

# Source. Note .dockerignore excludes test/, docs/, .git/, etc.
COPY . .

# Build flags:
#   -tags timetzdata  — embed Go's bundled tzdata so Asia/Dushanbe resolves.
#   -trimpath         — strip absolute build paths for reproducibility.
#   -ldflags "-s -w"  — strip debug info; smaller binary.
#   CGO_ENABLED=0     — pure-Go; modernc.org/sqlite is the only sqlite dep
#                       and it's CGO-free.
ENV CGO_ENABLED=0 GOOS=linux

RUN go build \
        -tags timetzdata \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/api \
        ./cmd/api

# ---------- runtime ----------
FROM alpine:3.20

# tzdata is technically redundant given -tags timetzdata, but it keeps the
# system TZ database available for any future tool that expects it on disk,
# and the cost is ~2MB. ca-certificates for any outbound HTTPS the binary
# might do later. tini reaps zombies in case any sub-tool is spawned.
RUN apk add --no-cache tzdata ca-certificates tini

# Non-root runtime user. UID 10001 to stay clear of any host-side conflicts.
RUN addgroup -S app && adduser -S -u 10001 -G app app

WORKDIR /app

# Binary.
COPY --from=builder /out/api /app/api

# Static assets the binary loads via relative ./internal/templates/ paths.
# See MIGRATION/00-inspection.md §B for the list of templates and the
# handlers that reference them.
COPY --from=builder /src/internal/templates /app/internal/templates

# Config template. The entrypoint will read .env-supplied values and write
# /app/configurations/App_dev.yaml here at container start. The committed
# App_dev.yaml in the source tree is shipped as the template (its structure
# is what we mirror at startup).
COPY --from=builder /src/configurations/App_dev.yaml /app/configurations/App_dev.yaml.tmpl

# Entrypoint script. Templates App_dev.yaml from env vars, then exec's CMD.
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

# Prepare storage subdirs the binary writes to during Excel import/export.
# These get covered by a named docker volume in docker-compose.yml so files
# persist across container restarts.
RUN mkdir -p \
        /app/storage/import_excel/input \
        /app/storage/import_excel/output \
        /app/storage/import_excel/return \
        /app/storage/import_excel/writeoff \
        /app/storage/import_excel/temp \
        /app/files && \
    chmod +x /usr/local/bin/docker-entrypoint.sh && \
    chown -R app:app /app

USER app

# Default to log-only permission enforcement (matches legacy no-op behavior).
# Operators flip this to "1" after post-migration permission verification.
ENV AUTH_PERMISSIONS_ENFORCE=0

# The binary binds to App.Host:App.Port. App.Host=0.0.0.0 is set by the
# entrypoint from APP_HOST in .env.backend; if absent it defaults to
# 127.0.0.1, which is wrong inside a Docker bridge — but the entrypoint
# defaults APP_HOST to 0.0.0.0 explicitly so this is safe.
EXPOSE 5000

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/docker-entrypoint.sh"]
CMD ["/app/api"]
