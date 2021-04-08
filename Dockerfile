FROM golang:alpine AS builder
RUN apk add git
RUN GO111MODULE=on go get -v github.com/miko/dmt

FROM alpine
ENTRYPOINT /bin/dmt
WORKDIR /dmt
COPY example /dmt/example
COPY --from=builder /go/bin/dmt /bin/dmt
