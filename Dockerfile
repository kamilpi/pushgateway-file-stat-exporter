# build stage
FROM golang:1.16.2 AS builder
# working directory
ENV GOPATH=/go/
WORKDIR /go/
COPY ./ ./src/
# rebuilt built in libraries and disabled cgo
WORKDIR /go/src/
RUN CGO_ENABLED=0 GOOS=linux go build -a -o file_stat .
# final stage
# FROM alpine:3.12.1
FROM busybox:1.33.0
# working directory
WORKDIR /app/
# copy the binary file into working directory
COPY --from=builder /go/src/file_stat ./bin/
# Run the backend command when the container starts.
CMD ["/app/bin/file_stat", "-run"]
