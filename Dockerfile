## BUILDER
FROM golang:1.22-bookworm AS builder

WORKDIR /src

COPY . .

RUN go build -o app ./cmd/listener


## DEPLOY
FROM debian:bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /cmd

COPY --from=builder /src/app /cmd/app

ENTRYPOINT ["/cmd/app"]
