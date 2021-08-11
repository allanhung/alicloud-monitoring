FROM golang:1.16 AS builder

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOPATH=/go
WORKDIR /go/src/github.com/allanhung/alicloud-monitoring

COPY cmd cmd/
COPY pkg pkg/
COPY main.go go.mod ./

RUN go mod tidy
RUN GOARCH=amd64 GOOS=linux go build -o /go/bin/alicloud-monitoring main.go

FROM debian:stretch-slim
RUN apt-get update && apt-get install -y git

ENV XDG_CONFIG_HOME=/opt

COPY --from=builder /go/bin/alicloud-monitoring /usr/bin/alicloud-monitoring

WORKDIR /working
ENTRYPOINT /usr/bin/alicloud-monitoring
