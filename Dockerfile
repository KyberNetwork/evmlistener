## BUILDER
FROM golang:1.21-bullseye as builder

WORKDIR /src

COPY . .

RUN go build -o app ./cmd/listener


## DEPLOY
FROM debian:bullseye

RUN apt-get update && \
    apt install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /cmd

COPY --from=builder /src/app /cmd/app

ENTRYPOINT /cmd/app
