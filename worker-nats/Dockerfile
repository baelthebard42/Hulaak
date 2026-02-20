# ---- Build Stage ----
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o worker ./worker.go


# ---- Runtime Stage ----
FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=builder /app/worker .


USER nonroot:nonroot


ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["./worker"]
