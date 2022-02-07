# builder
ARG BUILDER_BASE=golang:1.17.5-alpine3.15
FROM ${BUILDER_BASE} AS builder
ENV GOPROXY=https://goproxy.io,direct
WORKDIR /build
COPY . .
RUN go mod tidy
RUN go build ./cmd/...

# create image
FROM alpine:3.15
WORKDIR /
COPY --from=builder /build/st001 /st001
ENTRYPOINT ["/st001"]
EXPOSE 8080/tcp
