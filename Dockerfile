# ---- Stage 1: Build Vue SPA ----
FROM oven/bun:1-alpine AS node
WORKDIR /app
COPY dashboard/package.json dashboard/bun.lock ./
RUN bun install --frozen-lockfile
COPY dashboard/ .
RUN bun run vite build

# ---- Stage 2: Build Go binary ----
FROM golang:1.26-alpine AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN rm -rf dashboard/node_modules dashboard/public
COPY --from=node /app/dist ./dashboard/dist
RUN find dashboard/dist -type d -empty -delete
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o guardian .

# ---- Stage 3: Runtime ----
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=build /build/guardian /app/guardian
EXPOSE 8082
VOLUME ["/app/data"]
CMD ["/app/guardian"]