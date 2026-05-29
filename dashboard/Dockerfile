FROM oven/bun:1-alpine AS node
WORKDIR /app
COPY package.json ./
RUN bun install
COPY . .
RUN bun run vite build

FROM python:3.13-slim
COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv
WORKDIR /app
COPY api/pyproject.toml api/uv.lock ./
COPY api/firekeeper_dashboard/ ./firekeeper_dashboard/
RUN uv sync --frozen
COPY --from=node /app/dist ./dist
