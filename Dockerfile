FROM golang:1.16.3-buster as builder

WORKDIR /opt/app

COPY main.go ./
COPY go.mod ./

RUN go build -v -o signer


FROM registry.91.team/cryptopro/csp:latest

WORKDIR /opt/app

COPY --from=builder /opt/app/signer ./

RUN mkdir -p ./tmp

EXPOSE 3000

CMD ["./signer"]