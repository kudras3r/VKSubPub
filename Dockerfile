FROM golang:1.24.2 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o spapp ./cmd/subpub/main.go

FROM debian:stable-slim

WORKDIR /app

COPY --from=builder /app/spapp .

EXPOSE 44044

ENTRYPOINT ["./spapp"]