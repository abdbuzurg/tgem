# syntax=docker/dockerfile:1.7
#
# Multi-stage build for the tgem frontend (React 18 + Vite 4).
#
# Build: node:20-alpine. Runs `npm ci` + `npm run build`. The Vite build
# baked-in VITE_API_BASE_URL is the API base URL the resulting SPA will use
# for every request. Default /api makes the bundle origin-agnostic — it
# works under any host nginx (HTTP, HTTPS, any server_name) as long as
# /api/ is proxied to the backend on that same origin.
#
# Runtime: nginx:1.27-alpine. Serves the static dist with an internal nginx
# config that includes the SPA fallback (try_files … /index.html) and
# long-cache headers for hashed assets. Listens on :80 inside the
# container; the docker-compose port mapping exposes it as 127.0.0.1:8080
# on the host so the host nginx can reverse-proxy.

# ---------- builder ----------
FROM node:20-alpine AS builder

WORKDIR /src

# Dependency layer: only invalidated when package files change.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Build-time API base URL injection. Default /api (same-origin relative)
# means a single bundle works in any environment. Override with
# `docker compose build --build-arg VITE_API_BASE_URL=https://other.tld/api`
# if a non-relative URL is ever needed.
ARG VITE_API_BASE_URL=/api
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

# `tsc && vite build` per package.json. .dockerignore excludes the
# committed .env / .env.production so the ENV above is the sole source of
# the API URL — no precedence ambiguity.
RUN npm run build

# ---------- runtime ----------
FROM nginx:1.27-alpine

# Drop the default site config; our nginx.conf is the only server block.
RUN rm /etc/nginx/conf.d/default.conf

# Static assets.
COPY --from=builder /src/dist /usr/share/nginx/html

# Per-container nginx server block. SPA fallback + cache headers.
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

# The base image already sets the right CMD (nginx -g 'daemon off;') and
# runs the master as root, worker as the 'nginx' user. No overrides needed.
