# Stage 1: build frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci --prefer-offline
COPY frontend/ .
RUN npm run build

# Stage 2: build backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /qurt-quiz ./cmd/server

# Stage 3: minimal runtime image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /qurt-quiz /app/qurt-quiz
COPY --from=frontend-builder /app/frontend/dist /app/static
COPY packs/ /app/packs/

ENV ADDR=:8080
ENV PACK_DIR=/app/packs
ENV STATIC_DIR=/app/static

EXPOSE 8080
ENTRYPOINT ["/app/qurt-quiz"]
