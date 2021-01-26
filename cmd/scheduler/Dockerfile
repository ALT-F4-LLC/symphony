# Base stage

FROM golang:1.14-alpine as base_scheduler

RUN apk add --no-cache git protobuf

WORKDIR /go/src/github.com/erkrnt/symphony

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN go get -d -v ./...
RUN go install github.com/golang/protobuf/protoc-gen-go

COPY ./api ./api

RUN protoc --proto_path="$(pwd)" api/*.proto --go_out=plugins=grpc:../../../

COPY ./cmd/cli ./cmd/cli
COPY ./cmd/scheduler ./cmd/scheduler
COPY ./internal/scheduler ./internal/scheduler
COPY ./internal/cli ./internal/cli
COPY ./internal/service ./internal/service
COPY ./internal/utils ./internal/utils

RUN go install -v ./...

# Final stage

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /usr/local/bin

COPY --from=base_scheduler /go/bin/cli .
COPY --from=base_scheduler /go/bin/scheduler .

CMD [ "scheduler" ]
