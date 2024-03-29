FROM golang:alpine AS builder
RUN apk add git
ARG TAG=v0.2.30
RUN GOPRIVATE=github.com/miko/dmt GO111MODULE=on go install -v github.com/miko/dmt@${TAG}

FROM alpine
ENTRYPOINT /bin/dmt
WORKDIR /dmt
COPY example /dmt/example
COPY --from=builder /go/bin/dmt /bin/dmt
