# syntax=docker/dockerfile:1.7
# ----------------------------------------------------------------------
# LogiTrack frontend — multi-stage image
# Stage 1: build the Vue/Vite bundle with Node 22.
# Stage 2: serve the static assets with a minimal nginx image.
# Mission II contract: the runtime listens on :5174.
# ----------------------------------------------------------------------

FROM node:25-alpine AS build
WORKDIR /app

COPY package*.json ./
# Strict ci: with package-lock.json now committed (HIGH-10 follow-up
# from the 2026-04-27 audit), npm ci must succeed against the
# lockfile. We intentionally drop the previous `|| npm install`
# fallback so a lockfile drift fails the build loudly instead of
# silently re-resolving every ^-ranged dependency.
RUN npm ci --no-audit --no-fund

COPY . .
RUN npm run build

FROM nginx:1.27-alpine AS runtime
WORKDIR /usr/share/nginx/html
RUN rm -rf ./*

COPY --from=build /app/dist/ ./
COPY nginx.conf /etc/nginx/conf.d/default.conf

# nginx:1.27-alpine starts as root, drops to the "nginx" worker user via
# its own config. Forcing USER nginx here breaks /var/cache/nginx writes
# and the default entrypoint; rely on the upstream-baked privilege drop.

EXPOSE 5174
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget -qO- http://127.0.0.1:5174/ >/dev/null || exit 1

CMD ["nginx", "-g", "daemon off;"]
