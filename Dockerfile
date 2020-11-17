# Build the manager binary
FROM registry.cn-hangzhou.aliyuncs.com/knative-sample/golang:1.13-alpine3.10 as builder

# Copy in the go src
WORKDIR /go/src/github.com/knative-sample/event-display
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o event-display github.com/knative-sample/event-display/cmd

# Copy the event-display into a thin image
FROM alpine:3.7
WORKDIR /
COPY --from=builder /go/src/github.com/knative-sample/event-display/cmd app/
ENTRYPOINT ["/app/event-display"]