# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go build -o /out/carbonqt ./

FROM alpine:3.19
WORKDIR /app
RUN adduser -D -g '' carbonqt

COPY --from=builder /out/carbonqt /app/carbonqt

USER carbonqt
ENTRYPOINT ["/app/carbonqt"]