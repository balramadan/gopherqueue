FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go build -ldflags="-w -s" -o /app/bin/server ./cmd/server
RUN go build -ldflags="-w -s" -o /app/bin/worker ./cmd/worker
RUN go build -ldflags="-w -s" -o /app/bin/dashboard ./cmd/dashboard

FROM alpine:3.22 AS server
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/server .
EXPOSE 8080
ENTRYPOINT ["./server"]

FROM alpine:3.22 AS worker
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/worker .
ENTRYPOINT ["./worker"]

FROM alpine:3.22 AS dashboard
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/dashboard .
EXPOSE 3000
ENTRYPOINT ["./dashboard"]
