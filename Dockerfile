# ── Stage 1: Frontend build ──
FROM node:22-alpine AS frontend
WORKDIR /build
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ── Stage 2: Go build ──
FROM golang:1.26-alpine AS backend
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /build/dist ./internal/ui/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/craftctl ./cmd/main.go

# ── Stage 3: Runtime ──
FROM alpine:3.20
RUN apk add --no-cache openjdk21-jre-headless bash
ENV CRAFTCTL_HOME=/data
WORKDIR /app
COPY --from=backend /out/craftctl /usr/local/bin/craftctl
VOLUME ["/data"]
EXPOSE 8000
CMD ["craftctl"]
